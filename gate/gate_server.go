package gate

import (
	"encoding/binary"
	"fmt"
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/common/cust_error"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/network"
	"github.com/ambitiousmice/go-one/common/pktconn"
	"github.com/ambitiousmice/go-one/common/utils"
	"github.com/ambitiousmice/go-one/gate/dispatcher"
	"github.com/gorilla/websocket"
	"github.com/robfig/cron/v3"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/xtaci/kcp-go"
)

var gateServer *GateServer

var upgrader = websocket.Upgrader{
	ReadBufferSize:  consts.ClientProxyReadBufferSize,
	WriteBufferSize: consts.ClientProxyWriteBufferSize,
	WriteBufferPool: &sync.Pool{},
}

// GateServer implements the gate service logic
type GateServer struct {
	sync.RWMutex
	listenAddr                   string
	websocketListenAddr          string
	cron                         *cron.Cron
	tempClientProxies            map[string]*ClientProxy
	clientProxies                map[int64]*ClientProxy
	clientPacketQueue            chan *pktconn.Packet
	dispatcherClientPacketQueues []chan *pktconn.Packet

	status                  int32
	checkHeartbeatsInterval int
	clientTimeout           time.Duration

	NeedLogin    bool
	LoginManager LoginManager
}

func (gs *GateServer) String() string {
	return fmt.Sprintf("GateServer<%s>", gs.listenAddr)
}

func NewGateServer() *GateServer {
	if gateServer != nil {
		log.Panic("gate server only can be initialized once")
	}

	cronTab := cron.New(cron.WithSeconds())
	cronTab.Start()
	gateServer = &GateServer{
		tempClientProxies:       map[string]*ClientProxy{},
		clientProxies:           map[int64]*ClientProxy{},
		clientPacketQueue:       make(chan *pktconn.Packet, consts.GateServicePacketQueueSize),
		listenAddr:              gateConfig.Server.ListenAddr,
		websocketListenAddr:     gateConfig.Server.WebsocketListenAddr,
		cron:                    cronTab,
		NeedLogin:               gateConfig.Server.NeedLogin,
		checkHeartbeatsInterval: gateConfig.Server.HeartbeatCheckInterval,
		clientTimeout:           time.Second * time.Duration(gateConfig.Server.ClientTimeout),
	}

	dispatcherClientPacketQueues := make([]chan *pktconn.Packet, gateConfig.Server.DispatcherClientPacketQueuesSize)

	for i := 0; i < gateConfig.Server.DispatcherClientPacketQueuesSize; i++ {
		dispatcherClientPacketQueues[i] = make(chan *pktconn.Packet, consts.GateServicePacketQueueSize)
	}

	gateServer.dispatcherClientPacketQueues = dispatcherClientPacketQueues

	return gateServer
}

func (gs *GateServer) Run() {
	dispatcher.InitGameDispatchers(GetGateConfig().GameDispatcherConfigs, gs.dispatcherClientPacketQueues)

	go network.ServeTCPForever(gs.listenAddr, gs)
	go gs.serveKCP(gs.listenAddr)
	go network.ServeWebsocket(gs.websocketListenAddr, gs)

	runServerCronTask()

	setupSignals()

	gs.mainRoutine()
}

func runServerCronTask() {
	collectData := make(map[string]any)
	groupID, err := strconv.ParseInt(context.GetOneConfig().Nacos.Instance.GroupName, 10, 64)
	if err != nil {
		log.Errorf("gate groupName is not int: %s ,run failed", context.GetOneConfig().Nacos.Instance.GroupName)
		return
	}
	clusterId, err := strconv.ParseInt(context.GetOneConfig().Nacos.Instance.ClusterName, 10, 64)
	if err != nil {
		log.Errorf("gate ClusterName is not int: %s ,run failed", context.GetOneConfig().Nacos.Instance.ClusterName)
		return
	}
	collectData[consts.ServerName] = context.GetOneConfig().Nacos.Instance.Service
	collectData[consts.GroupId] = groupID
	collectData[consts.ClusterId] = clusterId

	gateServer.cron.AddFunc("@every 10s", func() {
		log.Infof("当前在线人数:%d", len(gateServer.clientProxies))
		log.Infof("客户端包队列长度:%d", len(gateServer.clientPacketQueue))
		dispatcherClientPacketQueueLength := 0
		for _, queue := range gateServer.dispatcherClientPacketQueues {
			dispatcherClientPacketQueueLength = dispatcherClientPacketQueueLength + len(queue)
		}
		log.Infof("分发客户端包长度:%d", dispatcherClientPacketQueueLength)

		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		totalMB := float64(stats.Sys) / 1024 / 1024
		log.Infof("Total Memory: %.2f MB", totalMB)
		memoryUsageMB := float64(stats.Sys-stats.HeapReleased) / 1024 / 1024
		log.Infof("Usage Memory: %.2f MB", memoryUsageMB)
	})

	gateServer.cron.AddFunc("@every 2s", func() {
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		totalMB := float64(stats.Sys) / 1024 / 1024
		memoryUsageMB := float64(stats.Sys-stats.HeapReleased) / 1024 / 1024

		collectData[consts.TotalMemory] = totalMB
		collectData[consts.UsageMemory] = memoryUsageMB
		collectData[consts.ConnectionCount] = len(gateServer.clientProxies)
		collectData[consts.Status] = gateServer.status
		collectData[consts.Metadata] = context.GetOneConfig().Nacos.Instance.Metadata
		resp := make(map[string]string)
		err := utils.Post(GetGateConfig().Params["monitorServerCollectDataUrl"].(string), collectData, &resp)
		if err != nil || resp["code"] != "0" {
			log.Warnf("上报数据失败:%s", err)
		}
	})
}

func setupSignals() {
	log.Infof("Setup signals ...")
	var signalChan = make(chan os.Signal, 1)
	signal.Ignore(syscall.Signal(10), syscall.Signal(12), syscall.SIGPIPE, syscall.SIGHUP)
	signal.Notify(signalChan, syscall.SIGTERM)

	go func() {
		for {
			sig := <-signalChan
			if sig == syscall.SIGTERM {

				gateServer.terminate()

				os.Exit(0)
			} else {
				log.Errorf("unexpected signal: %s", sig)
			}
		}
	}()
}

func (gs *GateServer) mainRoutine() {
	// 启动goroutine监听clientPacketQueue
	gs.status = consts.ServiceOnline
	go func() {
		for {
			select {
			case pkt := <-gs.clientPacketQueue:
				gs.handleClientProxyPacket(pkt)
				pkt.Release()
			}
		}
	}()

	// 启动goroutine监听dispatcherClientPacketQueues
	for _, queue := range gs.dispatcherClientPacketQueues {
		go func(q <-chan *pktconn.Packet) {
			for {
				select {
				case pkt := <-q:
					gs.handleDispatcherPacket(pkt)
					pkt.Release()
				}
			}
		}(queue)
	}

	// 阻塞方法，以防止退出
	select {}
}

// ServeTCPConnection handle TCP connections from clients
func (gs *GateServer) ServeTCPConnection(conn net.Conn) {
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetWriteBuffer(consts.ClientProxyWriteBufferSize)
	tcpConn.SetReadBuffer(consts.ClientProxyReadBufferSize)
	tcpConn.SetNoDelay(true)

	gs.handleClientConnection(conn)
}

func (gs *GateServer) ServeWebsocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("websocket upgrade error:", err)
		return
	}

	netConn := network.WebSocketConn{Conn: conn}
	log.Infof("websocket new connect request in")
	go gs.handleClientConnection(netConn)
}

func (gs *GateServer) serveKCP(addr string) {
	//kcpListener, err := kcp.ListenWithOptions(addr, nil, 10, 3) // fec 前向纠错
	kcpListener, err := kcp.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		log.Error(err)
	}

	log.Infof("Listening on KCP: %s ...", addr)

	for {
		conn, err := kcpListener.AcceptKCP()
		if err != nil {
			log.Error(err)
		}

		go gs.handleKCPConn(conn)
	}
}

func (gs *GateServer) handleKCPConn(conn *kcp.UDPSession) {
	log.Infof("KCP connection from %s", conn.RemoteAddr())

	conn.SetReadBuffer(consts.ClientProxyReadBufferSize)
	conn.SetWriteBuffer(consts.ClientProxyWriteBufferSize)
	conn.SetNoDelay(consts.KCP_NO_DELAY, consts.KCP_INTERNAL_UPDATE_TIMER_INTERVAL, consts.KCP_ENABLE_FAST_RESEND, consts.KCP_DISABLE_CONGESTION_CONTROL)
	conn.SetStreamMode(consts.KCP_SET_STREAM_MODE)
	conn.SetWriteDelay(consts.KCP_SET_WRITE_DELAY)
	conn.SetACKNoDelay(consts.KCP_SET_ACK_NO_DELAY)

	gs.handleClientConnection(conn)
}

func (gs *GateServer) handleClientConnection(conn net.Conn) {

	if gs.status != consts.ServiceOnline {
		conn.Close()
		return
	}

	cp := newClientProxy(conn)

	gs.addTempClientProxy(cp)

	if !gs.NeedLogin {
		cp.entityID = context.NextEntityID()
	}

	gs.addTempClientProxy(cp)

	jobID, err := cp.cron.AddFunc("@every 10s", func() {
		if cp.entityID == 0 {
			log.Infof("客户端:%s 未登录游戏,主动踢出", cp)
			cp.CloseAll()
		}
	})

	if err != nil {
		log.Errorf("客户端:%s 添加定时任务失败", cp)
		cp.CloseAll()
	}

	cp.cronMap[consts.CheckLogin] = jobID
	cp.SendConnectionSuccessFromServer()

	cp.cron.AddFunc("@every "+strconv.Itoa(gs.checkHeartbeatsInterval)+"s", func() {
		if time.Now().Sub(cp.heartbeatTime) > gs.clientTimeout {
			cp.HeartbeatTimeout()
		}
	})

	cp.serve()
}

func (gs *GateServer) onClientProxyClose(cp *ClientProxy) {
	gs.removeClientProxy(cp.entityID)
	gs.removeTempClientProxy(cp.clientID)

	log.Infof("%s: client %s disconnected", gs, cp)
}

// GetDispatcherClientPacketQueue handles packets received by dispatcher client
func (gs *GateServer) handleClientProxyPacket(pkt *pktconn.Packet) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("handle ClientProxy Packet error:%s", err)
		}
	}()

	cp := pkt.Src.Proxy.(*ClientProxy)
	cp.heartbeatTime = time.Now()
	//entityID := cp.entityID
	msgType := pkt.ReadUint16()

	log.Infof("收到客户端:%s 消息类型:%d", cp, msgType)
	switch msgType {
	case common_proto.GameMethodFromClient:
		err := cp.ForwardByDispatcher(pkt)
		if err != nil {
			customErr, ok := err.(*cust_error.CustomError)
			if ok {
				cp.SendMsg(common_proto.Error, &common_proto.ErrorResp{
					Code: customErr.ErrorCode,
				})
			} else {
				cp.SendMsg(common_proto.Error, &common_proto.ErrorResp{
					Code: common_proto.Game_Maintenance_Error,
				})
			}
		}
	case common_proto.HeartbeatFromClient:
		cp.SendHeartBeatAck()
	case common_proto.LoginFromClient:
		cp.Login(pkt)
	case common_proto.OfflineFromClient:
		cp.CloseAll()
	default:
		log.Errorf("unknown message type from client: %d", msgType)
	}

}

func (gs *GateServer) handleDispatcherPacket(packet *pktconn.Packet) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("handle Dispatcher Packet error:%s", err)
		}
	}()

	payload := packet.Payload()
	length := len(payload)
	entityID := int64(binary.LittleEndian.Uint64(payload[length-consts.EntityIDLength : length]))
	clientProxy := gs.getClientProxy(entityID)

	if clientProxy != nil {
		packet.ClearLastPayload(consts.EntityIDLength)
		err := clientProxy.Send(packet)
		if err != nil {
			log.Errorf("客户端:%s 发送消息失败:%s", clientProxy, err)
		}
	}
}

func (gs *GateServer) terminate() {
	if gs.status == consts.ServiceTerminating || gs.status == consts.ServiceTerminated {
		return
	}

	gs.status = consts.ServiceTerminating
	log.Infof("gate service terminating...")

	collectData := make(map[string]any)
	groupID, _ := strconv.ParseInt(context.GetOneConfig().Nacos.Instance.GroupName, 10, 64)
	clusterId, _ := strconv.ParseInt(context.GetOneConfig().Nacos.Instance.ClusterName, 10, 64)
	collectData[consts.ServerName] = context.GetOneConfig().Nacos.Instance.Service
	collectData[consts.GroupId] = groupID
	collectData[consts.ClusterId] = clusterId
	collectData[consts.Status] = gs.status
	resp := make(map[string]string)

	err := utils.Post(GetGateConfig().Params["monitorServerCollectDataUrl"].(string), collectData, &resp)
	if err != nil || resp["code"] != "0" {
		log.Warnf("停服上报数据失败:%s", err)
	}

	log.Infof("gate service terminating info report to monitor success")

	for _, cp := range gs.tempClientProxies {
		cp.CloseAll()
	}

	for _, cp := range gs.clientProxies {
		cp.CloseAll()
	}

	// TODO 处理通道信息

	gs.status = consts.ServiceTerminated

	collectData[consts.Status] = gs.status

	err = utils.Post(GetGateConfig().Params["monitorServerCollectDataUrl"].(string), collectData, &resp)
	if err != nil || resp["code"] != "0" {
		log.Warnf("停服上报数据失败:%s", err)
	}

	log.Infof("gate service terminated info report to monitor success")

	log.Infof("gate service terminated")
}

func (gs *GateServer) addTempClientProxy(cp *ClientProxy) {
	gs.Lock()
	gs.tempClientProxies[cp.clientID] = cp
	gs.Unlock()
}

func (gs *GateServer) removeTempClientProxy(clientID string) {
	gs.Lock()
	delete(gs.tempClientProxies, clientID)
	gs.Unlock()
}

func (gs *GateServer) getTempClientProxy(clientID string) *ClientProxy {
	gs.RLock()
	defer gs.RUnlock()
	return gs.tempClientProxies[clientID]

}

func (gs *GateServer) addClientProxy(cp *ClientProxy) {
	gs.Lock()
	gs.clientProxies[cp.entityID] = cp
	gs.Unlock()
}

func (gs *GateServer) removeClientProxy(entityID int64) {
	gs.Lock()
	delete(gs.clientProxies, entityID)
	gs.Unlock()
}

func (gs *GateServer) getClientProxy(entityID int64) *ClientProxy {
	gs.RLock()
	defer gs.RUnlock()
	return gs.clientProxies[entityID]

}

func (gs *GateServer) Broadcast(msg any) {
	gs.RLock()
	defer gs.RUnlock()
	for _, proxy := range gs.clientProxies {
		proxy.SendMsg(common_proto.BroadcastFromServer, msg)
	}
}
func GetGateServer() *GateServer {
	return gateServer
}

package game

import (
	"fmt"
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/network"
	"github.com/ambitiousmice/go-one/common/pktconn"
	"github.com/ambitiousmice/go-one/common/pool/goroutine_pool"
	"github.com/ambitiousmice/go-one/common/utils"
	"github.com/ambitiousmice/go-one/game/entity"
	"github.com/ambitiousmice/go-one/game/processor_center"
	"github.com/ambitiousmice/go-one/game/proxy"
	"github.com/robfig/cron/v3"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var gameServer *GameServer

// GameServer implements the Game service logic
type GameServer struct {
	sync.Mutex
	Game       string
	listenAddr string
	cron       *cron.Cron

	GpMutex         sync.RWMutex
	gateProxies     map[string]*proxy.GateProxy
	gateNodeProxies map[uint8][]*proxy.GateProxy
	pollingIndex    uint64

	GatePacketQueue chan *pktconn.Packet

	status                  uint8
	checkHeartbeatsInterval int
	gateTimeout             time.Duration
	cronTaskMap             map[string]cron.EntryID
}

func NewGameServer() *GameServer {
	if gameServer != nil {
		log.Panic("Game server only can be initialized once")
	}

	crontab := cron.New(cron.WithSeconds())
	crontab.Start()
	gameServer = &GameServer{
		gateProxies:     map[string]*proxy.GateProxy{},
		gateNodeProxies: map[uint8][]*proxy.GateProxy{},
		GatePacketQueue: make(chan *pktconn.Packet, consts.GameServicePacketQueueSize),

		Game:                    gameConfig.Server.Game,
		listenAddr:              gameConfig.Server.ListenAddr,
		cron:                    crontab,
		checkHeartbeatsInterval: gameConfig.Server.HeartbeatCheckInterval,
		gateTimeout:             time.Second * time.Duration(gameConfig.Server.GateTimeout),
		cronTaskMap:             map[string]cron.EntryID{},
	}

	if len(gameConfig.SceneManagerConfigs) == 0 {
		log.Warnf("no scene manager config,will only support lobby scene")
	} else {
		for _, config := range gameConfig.SceneManagerConfigs {
			entity.SceneManagerContext[config.SceneType] = entity.NewSceneManager(config.SceneType,
				config.SceneMaxPlayerNum,
				config.SceneIDStart,
				config.SceneIDEnd,
				config.MatchStrategy,
				config.EnableAOI,
				config.AOIDistance,
				time.Duration(config.TickRate)*time.Millisecond,
			)
		}
	}

	entity.SetGameServer(gameServer)

	return gameServer
}

func (gs *GameServer) Run() {
	go network.ServeTCPForever(gs.listenAddr, gs)
	log.Infof("心跳检测间隔:%ds,网关超时时间:%fs", gs.checkHeartbeatsInterval, gs.gateTimeout.Seconds())

	runServerCronTask()

	setupSignals()

	gs.mainRoutine()
}

func runServerCronTask() {
	collectData := make(map[string]any)
	groupID, err := strconv.ParseInt(context.GetOneConfig().Nacos.Instance.GroupName, 10, 64)
	if err != nil {
		log.Errorf("game groupName is not int: %s ,run failed", context.GetOneConfig().Nacos.Instance.GroupName)
		return
	}
	clusterId, err := strconv.ParseInt(context.GetOneConfig().Nacos.Instance.ClusterName, 10, 64)
	if err != nil {
		log.Errorf("game ClusterName is not int: %s ,run failed", context.GetOneConfig().Nacos.Instance.ClusterName)
		return
	}
	collectData[consts.ServerName] = context.GetOneConfig().Nacos.Instance.Service
	collectData[consts.GroupId] = groupID
	collectData[consts.ClusterId] = clusterId

	gameServer.cron.AddFunc("@every 10s", func() {
		log.Infof("当前链接数:%d", len(gameServer.gateProxies))
		log.Infof("网关包队列长度:%d", len(gameServer.GatePacketQueue))
		log.Infof("场景消息队列长度:%d", entity.GetSceneMsgChannelSize())
		log.Infof("当前在线人数:%d", entity.GetPlayerCount())

		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		totalMB := float64(stats.Sys) / 1024 / 1024
		log.Infof("Total Memory: %.2f MB", totalMB)
		memoryUsageMB := float64(stats.Sys-stats.HeapReleased) / 1024 / 1024
		log.Infof("Usage Memory: %.2f MB", memoryUsageMB)
	})

	gameServer.cron.AddFunc("@every 2s", func() {
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		totalMB := float64(stats.Sys) / 1024 / 1024
		memoryUsageMB := float64(stats.Sys-stats.HeapReleased) / 1024 / 1024

		collectData[consts.TotalMemory] = totalMB
		collectData[consts.UsageMemory] = memoryUsageMB
		collectData[consts.ConnectionCount] = entity.GetPlayerCount()
		collectData[consts.Metadata] = context.GetOneConfig().Nacos.Instance.Metadata
		resp := make(map[string]string)
		err := utils.Post(GetGameConfig().Params["monitorServerCollectDataUrl"].(string), collectData, &resp)
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
				gameServer.terminate()
				os.Exit(0)
			} else {
				log.Errorf("unexpected signal: %s", sig)
			}
		}
	}()
}

func (gs *GameServer) mainRoutine() {
	for {
		select {
		case pkt := <-gs.GatePacketQueue:
			err := goroutine_pool.Submit(func() {
				gs.handleGatePacket(pkt)
				pkt.Release()
			})

			if err != nil {
				log.Errorf("submit GatePacket task error:%s", err.Error())
			}
		}
	}
}

func (gs *GameServer) String() string {
	return fmt.Sprintf("GameServer<%s>", gs.listenAddr)
}

func GetGameServer() *GameServer {
	return gameServer
}

// ServeTCPConnection handle TCP connections from clients
func (gs *GameServer) ServeTCPConnection(conn net.Conn) {
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetWriteBuffer(consts.GateProxyWriteBufferSize)
	tcpConn.SetReadBuffer(consts.GateProxyReadBufferSize)
	tcpConn.SetNoDelay(true)

	gs.handleClientConnection(conn)
}

func (gs *GameServer) handleClientConnection(conn net.Conn) {

	if gs.status == consts.ServiceTerminating {
		conn.Close()
		return
	}

	cp := proxy.NewClientProxy(conn, gs, processor_center.GPM)

	cp.Cron.AddFunc("@every "+strconv.Itoa(gs.checkHeartbeatsInterval)+"s", func() {
		if time.Now().Sub(cp.HeartbeatTime) > gs.gateTimeout {
			log.Infof("网关:%s 心跳检测超时", cp)
			cp.CloseAll()
		}
	})

	cp.Serve(gs.GatePacketQueue)
}

func (gs *GameServer) OnClientProxyClose(cp *proxy.GateProxy) {
	clientProxy := gs.getGateProxy(cp.ProxyID)
	if clientProxy == nil {
		return
	}
	gs.removeGateProxy(cp)

	log.Infof("client %s disconnected", cp)
}

// GetDispatcherClientPacketQueue handles packets received by dispatcher client
func (gs *GameServer) handleGatePacket(pkt *pktconn.Packet) {

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("handle gate packet error,Recover from panic: %v", r)
		}
	}()

	gp := pkt.Src.Proxy.(*proxy.GateProxy)
	gp.HeartbeatTime = time.Now()
	msgType := pkt.ReadUint16()

	//log.Infof("receive msg:%s 消息类型:%d", gp, msgType)

	switch msgType {
	case common_proto.GameMethodFromClient:
		gp.HandleGameLogic(pkt)
	case common_proto.HeartbeatFromDispatcher:
		gp.SendHeartBeatAck()
	case common_proto.GameDispatcherChannelInfoFromDispatcher:
		gp.Handle3002(pkt)
	case common_proto.NewPlayerConnectionFromDispatcher:
		gp.Handle3003(pkt)
	case common_proto.PlayerDisconnectedFromDispatcher:
		gp.Handle3004(pkt)
	default:
		log.Errorf("unknown message type from client: %d", msgType)
	}

}

func (gs *GameServer) terminate() {
	if gs.status == consts.ServiceTerminating || gs.status == consts.ServiceTerminated {
		return
	}

	gs.status = consts.ServiceTerminating
	log.Infof("game service terminating...")

	collectData := make(map[string]any)
	groupID, _ := strconv.ParseInt(context.GetOneConfig().Nacos.Instance.GroupName, 10, 64)
	clusterId, _ := strconv.ParseInt(context.GetOneConfig().Nacos.Instance.ClusterName, 10, 64)
	collectData[consts.ServerName] = context.GetOneConfig().Nacos.Instance.Service
	collectData[consts.GroupId] = groupID
	collectData[consts.ClusterId] = clusterId
	collectData[consts.Status] = gs.status
	resp := make(map[string]string)

	err := utils.Post(GetGameConfig().Params["monitorServerCollectDataUrl"].(string), collectData, &resp)
	if err != nil || resp["code"] != "0" {
		log.Warnf("停服上报数据失败:%s", err)
	}

	log.Infof("game service terminating info report to monitor success")

	for _, cp := range gs.gateProxies {
		cp.CloseAll()
	}

	gs.status = consts.ServiceTerminated

	// TODO 处理通道信息

	gs.status = consts.ServiceTerminated

	collectData[consts.Status] = gs.status

	err = utils.Post(GetGameConfig().Params["monitorServerCollectDataUrl"].(string), collectData, &resp)
	if err != nil || resp["code"] != "0" {
		log.Warnf("停服上报数据失败:%s", err)
	}

	log.Infof("game service terminated info report to monitor success")

	log.Infof("game service terminated")
}

func (gs *GameServer) getGateProxy(proxyID string) *proxy.GateProxy {
	gs.GpMutex.RLock()
	defer gs.GpMutex.RUnlock()
	return gs.gateProxies[proxyID]
}

func (gs *GameServer) AddGateProxy(gp *proxy.GateProxy) {
	gs.GpMutex.Lock()
	defer gs.GpMutex.Unlock()

	for _, p := range gs.gateProxies {
		if p.GateClusterID == gp.GateClusterID && p.DispatcherChannelID == gp.DispatcherChannelID {
			gp.CloseAll()
			break
		}
	}
	gs.gateProxies[gp.ProxyID] = gp
	nodeProxies := gs.gateNodeProxies[gp.GateClusterID]
	if nodeProxies == nil {
		nodeProxies = []*proxy.GateProxy{gp}
		gs.gateNodeProxies[gp.GateClusterID] = nodeProxies
	} else {
		nodeProxies = append(nodeProxies, gp)
		gs.gateNodeProxies[gp.GateClusterID] = nodeProxies
	}
}

func (gs *GameServer) removeGateProxy(cp *proxy.GateProxy) {
	gs.GpMutex.Lock()
	defer gs.GpMutex.Unlock()
	delete(gs.gateProxies, cp.ProxyID)

	nodeProxies := gs.gateNodeProxies[cp.GateClusterID]
	if nodeProxies == nil {
		return
	}
	for i, proxy := range nodeProxies {
		if proxy.DispatcherChannelID == cp.DispatcherChannelID {
			gs.gateNodeProxies[cp.GateClusterID] = append(nodeProxies[:i], nodeProxies[i+1:]...)
			break
		}
	}
}

func (gs *GameServer) GetGateProxyByGateClusterID(gateClusterID uint8) *proxy.GateProxy {
	gs.GpMutex.RLock()
	defer gs.GpMutex.RUnlock()

	nodeProxies := gs.gateNodeProxies[gateClusterID]
	if len(nodeProxies) == 0 {
		return nil
	}

	pollingIndex := atomic.AddUint64(&gs.pollingIndex, 1)

	nextIndex := pollingIndex % uint64(len(nodeProxies))

	gateProxy := nodeProxies[nextIndex]

	return gateProxy
}

func (gs *GameServer) SendAndRelease(gateClusterID uint8, packet *pktconn.Packet) {
	gateProxy := gs.GetGateProxyByGateClusterID(gateClusterID)
	if gateProxy == nil {
		log.Errorf("not found gate proxy:%d", gateClusterID)
		packet.Release()
		return
	}

	err := gateProxy.SendAndRelease(packet)

	if err != nil {
		log.Errorf("%s send Game msg error: %s", gateProxy, err)
	}
}

func (gs *GameServer) AddCronTask(taskName string, spec string, method func()) error {
	taskID := gs.cronTaskMap[taskName]
	if taskID != 0 {
		delete(gs.cronTaskMap, strconv.Itoa(int(taskID)))
	}

	newTaskID, err := gs.cron.AddFunc(spec, method)
	if err != nil {
		return err
	}

	gs.cronTaskMap[taskName] = newTaskID

	return nil
}

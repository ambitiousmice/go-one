package gate

import (
	"encoding/binary"
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/consts"
	"go-one/common/context"
	"go-one/common/log"
	"go-one/common/network"
	"go-one/common/pktconn"
	"go-one/common/proto"
	"go-one/gate/dispatcher"
	"golang.org/x/net/websocket"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/xtaci/kcp-go"
)

var gateServer *GateServer

// GateServer implements the gate service logic
type GateServer struct {
	sync.RWMutex
	listenAddr                  string
	cron                        *cron.Cron
	tempClientProxies           map[string]*ClientProxy
	clientProxies               map[int64]*ClientProxy
	clientPacketQueue           chan *pktconn.Packet
	dispatcherClientPacketQueue chan *pktconn.Packet

	status                  uint8
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
		panic("gate server only can be initialized once")
	}

	cron := cron.New(cron.WithSeconds())
	cron.Start()
	gateServer = &GateServer{
		tempClientProxies:           map[string]*ClientProxy{},
		clientProxies:               map[int64]*ClientProxy{},
		clientPacketQueue:           make(chan *pktconn.Packet, consts.GateServicePacketQueueSize),
		dispatcherClientPacketQueue: make(chan *pktconn.Packet, consts.GateServicePacketQueueSize),
		listenAddr:                  gateConfig.Server.ListenAddr,
		cron:                        cron,
		NeedLogin:                   gateConfig.Server.NeedLogin,
		checkHeartbeatsInterval:     gateConfig.Server.HeartbeatCheckInterval,
		clientTimeout:               time.Second * time.Duration(gateConfig.Server.ClientTimeout),
	}

	return gateServer
}

func (gs *GateServer) Run() {

	dispatcher.InitGameDispatchers(GetGateConfig().GameDispatcherConfigs, gs.dispatcherClientPacketQueue)

	go network.ServeTCPForever(gs.listenAddr, gs)
	go gs.serveKCP(gs.listenAddr)

	log.Infof("心跳检测间隔:%ds,客户端超时时间:%fs", gs.checkHeartbeatsInterval, gs.clientTimeout.Seconds())

	gs.cron.AddFunc("@every 20s", func() {
		log.Infof("当前在线人数:%d", len(gs.clientProxies))
		log.Infof("客户端包队列长度:%d", len(gs.clientPacketQueue))
		log.Infof("分发客户端包长度:%d", len(gs.dispatcherClientPacketQueue))
	})

	gs.mainRoutine()
}

func (gs *GateServer) mainRoutine() {
	for {
		select {
		case pkt := <-gs.clientPacketQueue:
			go func() {
				gs.handleClientProxyPacket(pkt)
				pkt.Release()
			}()

		case pkt := <-gs.dispatcherClientPacketQueue:
			go func() {
				gs.handleDispatcherPacket(pkt)
				pkt.Release()
			}()
		}
	}
}

// ServeTCPConnection handle TCP connections from clients
func (gs *GateServer) ServeTCPConnection(conn net.Conn) {
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetWriteBuffer(consts.ClientProxyWriteBufferSize)
	tcpConn.SetReadBuffer(consts.ClientProxyReadBufferSize)
	tcpConn.SetNoDelay(true)

	gs.handleClientConnection(conn)
}

func (gs *GateServer) serveKCP(addr string) {
	//kcpListener, err := kcp.ListenWithOptions(addr, nil, 10, 3) // fec 前向纠错
	kcpListener, err := kcp.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		log.Panic(err)
	}

	log.Infof("Listening on KCP: %s ...", addr)

	for {
		conn, err := kcpListener.AcceptKCP()
		if err != nil {
			log.Panic(err)
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

func (gs *GateServer) handleWebSocketConn(wsConn *websocket.Conn) {
	log.Infof("WebSocket Connection: %s", wsConn.RemoteAddr())
	wsConn.PayloadType = websocket.BinaryFrame
	gs.handleClientConnection(wsConn)
}

func (gs *GateServer) handleClientConnection(conn net.Conn) {

	if gs.status == consts.ServiceTerminating {
		conn.Close()
		return
	}

	cp := newClientProxy(conn)

	gs.addTempClientProxy(cp)

	if !gs.NeedLogin {
		cp.entityID = context.NextEntityID()
	}

	gs.addTempClientProxy(cp)

	jobID, err := cp.cron.AddFunc("@every 3s", func() {
		if cp.entityID == 0 {
			log.Infof("客户端:%s 未登录游戏,主动踢出", cp)
			cp.CloseAll()
		}
	})

	if err != nil {
		log.Errorf("客户端:%s 添加定时任务失败", cp)
		cp.CloseAll()
	}

	cp.cronMap[consts.CheckEnterGame] = jobID
	cp.SendEnterGameFromServer()

	cp.cron.AddFunc("@every "+strconv.Itoa(gs.checkHeartbeatsInterval)+"s", func() {
		if time.Now().Sub(cp.heartbeatTime) > gs.clientTimeout {
			cp.HeartbeatTimeout()
		}
	})

	cp.serve()
}

func (gs *GateServer) onClientProxyClose(cp *ClientProxy) {
	clientProxy := gs.getTempClientProxy(cp.clientID)
	if clientProxy == nil {
		clientProxy = gs.getClientProxy(cp.entityID)
		if clientProxy == nil {
			return
		}
		gs.removeClientProxy(clientProxy.entityID)
	}
	gs.removeTempClientProxy(clientProxy.clientID)

	log.Infof("%s: client %s disconnected", gs, cp)
}

// GetDispatcherClientPacketQueue handles packets received by dispatcher client
func (gs *GateServer) handleClientProxyPacket(pkt *pktconn.Packet) {
	cp := pkt.Src.Proxy.(*ClientProxy)
	cp.heartbeatTime = time.Now()
	//entityID := cp.entityID
	msgType := pkt.ReadUint16()

	//log.Infof("收到客户端:%s 消息类型:%d", cp, msgType)
	switch msgType {
	case proto.GameMethodFromClient:
		cp.ForwardByDispatcher(pkt)
	case proto.HeartbeatFromClient:
		cp.SendHeartBeatAck()
	case proto.EnterGameFromClient:
		cp.EnterGame(pkt)
	case proto.OfflineFromClient:
		cp.CloseAll()
	default:
		log.Panicf("unknown message type from client: %d", msgType)
	}

}

func (gs *GateServer) handleDispatcherPacket(packet *pktconn.Packet) {
	payload := packet.Payload()
	length := len(payload)
	entityID := int64(binary.LittleEndian.Uint64(payload[length-consts.EntityIDLength : length]))

	clientProxy := gs.getClientProxy(entityID)

	if clientProxy != nil {
		err := clientProxy.Send(packet)
		if err != nil {
			log.Errorf("客户端:%s 发送消息失败:%s", clientProxy, err)
		}
	}
}

func (gs *GateServer) terminate() {
	gs.status = consts.ServiceTerminating

	for _, cp := range gs.tempClientProxies {
		cp.CloseAll()
	}

	for _, cp := range gs.clientProxies {
		cp.CloseAll()
	}

	gs.status = consts.ServiceTerminated
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

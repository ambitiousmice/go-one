package game

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/consts"
	"go-one/common/log"
	"go-one/common/network"
	"go-one/common/pktconn"
	"go-one/common/proto"
	"net"
	"strconv"
	"sync"
	"time"
)

var gameServer *GameServer

// GameServer implements the game service logic
type GameServer struct {
	sync.Mutex
	listenAddr      string
	cron            *cron.Cron
	gateProxies     map[string]*GateProxy
	gatePacketQueue chan *pktconn.Packet

	status                  uint8
	checkHeartbeatsInterval int
	gateTimeout             time.Duration
}

func NewGameServer() *GameServer {
	if gameServer != nil {
		panic("game server only can be initialized once")
	}

	cron := cron.New(cron.WithSeconds())
	cron.Start()
	gameServer = &GameServer{
		gateProxies:             map[string]*GateProxy{},
		gatePacketQueue:         make(chan *pktconn.Packet, consts.GameServicePacketQueueSize),
		listenAddr:              gameConfig.Server.ListenAddr,
		cron:                    cron,
		checkHeartbeatsInterval: gameConfig.Server.HeartbeatCheckInterval,
		gateTimeout:             time.Second * time.Duration(gameConfig.Server.GateTimeout),
	}

	return gameServer
}

func (gs *GameServer) Run() {
	go network.ServeTCPForever(gs.listenAddr, gs)
	log.Infof("心跳检测间隔:%ds,网关超时时间:%fs", gs.checkHeartbeatsInterval, gs.gateTimeout.Seconds())

	gs.cron.AddFunc("@every 20s", func() {
		log.Infof("当前链接数:%d", len(gs.gateProxies))
		log.Infof("网关包队列长度:%d", len(gs.gatePacketQueue))
	})

	gs.mainRoutine()
}

func (gs *GameServer) mainRoutine() {
	for {
		select {
		case pkt := <-gs.gatePacketQueue:
			go func() {
				gs.handleGatePacket(pkt)
				pkt.Release()
			}()
		}
	}
}

func (gs *GameServer) String() string {
	return fmt.Sprintf("GameServer<%s>", gs.listenAddr)
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

	cp := newClientProxy(conn)
	gs.gateProxies[cp.proxyID] = cp

	cp.cron.AddFunc("@every "+strconv.Itoa(gs.checkHeartbeatsInterval)+"s", func() {
		if time.Now().Sub(cp.heartbeatTime) > gs.gateTimeout {
			log.Infof("网关:%s 心跳检测超时", cp)
			cp.CloseAll()
		}
	})

	cp.serve()
}

func (gs *GameServer) onClientProxyClose(cp *GateProxy) {
	clientProxy := gs.gateProxies[cp.proxyID]
	if clientProxy == nil {
		return
	}
	delete(gs.gateProxies, cp.proxyID)

	log.Infof("client %s disconnected", cp)
}

// GetDispatcherClientPacketQueue handles packets received by dispatcher client
func (gs *GameServer) handleGatePacket(pkt *pktconn.Packet) {

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recover from panic: %v\n", r)
		}
	}()

	cp := pkt.Src.Proxy.(*GateProxy)
	cp.heartbeatTime = time.Now()
	//entityID := cp.entityID
	msgType := pkt.ReadUint16()

	//log.Infof("receive msg:%s 消息类型:%d", cp, msgType)

	switch msgType {
	case proto.GameMethodFromClient:
		gameReq := &proto.GameReq{}
		pkt.ReadData(gameReq)
		entityID := pkt.ReadInt64()
		gameProcess(entityID, gameReq)
	case proto.HeartbeatFromDispatcher:
		cp.SendHeartBeatAck()
	case proto.OfflineFromClient:
		cp.CloseAll()
	case proto.GameDispatcherChannelInfoFromDispatcher:
		cp.handle3002(pkt)
	case proto.NewPlayerConnectionFromDispatcher:
		cp.handle3003(pkt)
	case proto.PlayerDisconnectedFromDispatcher:
		cp.handle3004(pkt)
	default:
		log.Errorf("unknown message type from client: %d", msgType)
	}

}

func (gs *GameServer) terminate() {
	gs.status = consts.ServiceTerminating

	for _, cp := range gs.gateProxies {
		cp.CloseAll()
	}

	gs.status = consts.ServiceTerminated
}

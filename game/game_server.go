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
	listenAddr string
	cron       *cron.Cron

	gpMutex         sync.RWMutex
	gateProxies     map[string]*GateProxy
	gateNodeProxies map[uint8][]*GateProxy
	pollingIndex    uint8

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
		gateNodeProxies:         map[uint8][]*GateProxy{},
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

	cp.cron.AddFunc("@every "+strconv.Itoa(gs.checkHeartbeatsInterval)+"s", func() {
		if time.Now().Sub(cp.heartbeatTime) > gs.gateTimeout {
			log.Infof("网关:%s 心跳检测超时", cp)
			cp.CloseAll()
		}
	})

	cp.serve()
}

func (gs *GameServer) onClientProxyClose(cp *GateProxy) {
	clientProxy := gs.getGateProxy(cp.proxyID)
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
			log.Errorf("Recover from panic: %v\n", r)
		}
	}()

	gp := pkt.Src.Proxy.(*GateProxy)
	gp.heartbeatTime = time.Now()
	msgType := pkt.ReadUint16()

	//log.Infof("receive msg:%s 消息类型:%d", gp, msgType)

	switch msgType {
	case proto.GameMethodFromClient:
		gp.handleGameLogic(pkt)
	case proto.HeartbeatFromDispatcher:
		gp.SendHeartBeatAck()
	case proto.OfflineFromClient:
		gp.CloseAll()
	case proto.GameDispatcherChannelInfoFromDispatcher:
		gp.handle3002(pkt)
	case proto.NewPlayerConnectionFromDispatcher:
		gp.handle3003(pkt)
	case proto.PlayerDisconnectedFromDispatcher:
		gp.handle3004(pkt)
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

func (gs *GameServer) getGateProxy(proxyID string) *GateProxy {
	gs.gpMutex.RLock()
	defer gs.gpMutex.RUnlock()
	return gs.gateProxies[proxyID]
}

func (gs *GameServer) addGateProxy(cp *GateProxy) {
	gs.gpMutex.Lock()
	defer gs.gpMutex.Unlock()
	gs.gateProxies[cp.proxyID] = cp
	nodeProxies := gs.gateNodeProxies[cp.gateID]
	if nodeProxies == nil {
		nodeProxies = []*GateProxy{cp}
		gs.gateNodeProxies[cp.gateID] = nodeProxies
	} else {
		nodeProxies = append(nodeProxies, cp)
		gs.gateNodeProxies[cp.gateID] = nodeProxies
	}
}

func (gs *GameServer) removeGateProxy(cp *GateProxy) {
	gs.gpMutex.Lock()
	defer gs.gpMutex.Unlock()
	delete(gs.gateProxies, cp.proxyID)

	nodeProxies := gs.gateNodeProxies[cp.gateID]
	if nodeProxies == nil {
		return
	}
	for i, proxy := range nodeProxies {
		if proxy.dispatcherChannelID == cp.dispatcherChannelID {
			gs.gateNodeProxies[cp.gateID] = append(nodeProxies[:i], nodeProxies[i+1:]...)
			break
		}
	}
}

func (gs *GameServer) getGateProxyByGateID(gateID uint8) *GateProxy {
	gs.gpMutex.Lock()
	defer gs.gpMutex.Unlock()
	nodeProxies := gs.gateNodeProxies[gateID]
	if nodeProxies == nil {
		return nil
	}

	gs.pollingIndex++
	if gs.pollingIndex >= uint8(len(nodeProxies)) {
		gs.pollingIndex = 0
	}
	gateProxy := nodeProxies[gs.pollingIndex]
	return gateProxy
}

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
	"reflect"
	"strconv"
	"sync"
	"time"
)

var gameServer *GameServer

// GameServer implements the Game service logic
type GameServer struct {
	sync.Mutex
	Game       string
	listenAddr string
	cron       *cron.Cron

	gpMutex         sync.RWMutex
	gateProxies     map[string]*GateProxy
	gateNodeProxies map[uint8][]*GateProxy
	pollingIndex    uint8

	smMutex       sync.RWMutex
	SceneManagers map[string]*SceneManager
	sceneTypes    map[string]reflect.Type

	gatePacketQueue chan *pktconn.Packet

	status                  uint8
	checkHeartbeatsInterval int
	gateTimeout             time.Duration
}

func NewGameServer() *GameServer {
	if gameServer != nil {
		panic("Game server only can be initialized once")
	}

	crontab := cron.New(cron.WithSeconds())
	crontab.Start()
	gameServer = &GameServer{
		gateProxies:             map[string]*GateProxy{},
		gateNodeProxies:         map[uint8][]*GateProxy{},
		SceneManagers:           map[string]*SceneManager{},
		sceneTypes:              map[string]reflect.Type{},
		gatePacketQueue:         make(chan *pktconn.Packet, consts.GameServicePacketQueueSize),
		Game:                    gameConfig.Server.Game,
		listenAddr:              gameConfig.Server.ListenAddr,
		cron:                    crontab,
		checkHeartbeatsInterval: gameConfig.Server.HeartbeatCheckInterval,
		gateTimeout:             time.Second * time.Duration(gameConfig.Server.GateTimeout),
	}

	for _, config := range gameConfig.SceneManagerConfigs {
		gameServer.SceneManagers[config.SceneType] = NewSceneManager(config.SceneType, config.SceneMaxPlayerNum, config.SceneIDStart, config.SceneIDEnd, config.MatchStrategy)
	}

	gameServer.RegisterRoomType(&SceneLobby{})

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
			log.Panicf("handle gate packet error,Recover from panic: %v\n", r)
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

// RegisterRoomType register a room type
func (gs *GameServer) RegisterRoomType(room IScene) {
	if gs.sceneTypes[room.GetSceneType()] != nil {
		panic("room type already registered, sceneType:" + room.GetSceneType())
	}

	objVal := reflect.ValueOf(room)
	objType := objVal.Type()

	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	gs.sceneTypes[room.GetSceneType()] = objType
}

func (gs *GameServer) getSceneObjType(sceneType string) reflect.Type {
	objType := gs.sceneTypes[sceneType]
	if objType == nil {
		panic("scene type not found, sceneType:" + sceneType)
	}

	return objType
}

func (gs *GameServer) GetSceneManager(sceneType string) *SceneManager {
	sceneManager := gs.SceneManagers[sceneType]

	if sceneManager == nil {
		panic("scene manager not found, sceneType:" + sceneType)
	}

	return sceneManager
}

func (gs *GameServer) JoinScene(sceneType string, player *Player) {
	sceneManager := gs.GetSceneManager(sceneType)

	scene := sceneManager.GetSceneByStrategy()
	if scene == nil {
		player.SendCommonErrorMsg(ServerIsFull)
	}
	scene.Join(player)
}

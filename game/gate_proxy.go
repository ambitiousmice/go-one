package game

import (
	context2 "context"
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/consts"
	"go-one/common/context"
	"go-one/common/log"
	"go-one/common/pktconn"
	"go-one/common/proto"
	"net"
	"time"
)

// GateProxy is a Game client connections managed by gate
type GateProxy struct {
	*pktconn.PacketConn
	proxyID             string
	dispatcherChannelID uint8
	gateID              uint8

	heartbeatTime time.Time
	cron          *cron.Cron
	cronMap       map[string]cron.EntryID
}

func newClientProxy(_conn net.Conn) *GateProxy {

	_conn = pktconn.NewBufferedConn(_conn, consts.BufferedReadBufferSize, consts.BufferedWriteBufferSize)
	gateProxy := &GateProxy{
		proxyID:       context.NextClientID(),
		heartbeatTime: time.Now(),
		cron:          cron.New(cron.WithSeconds()),
		cronMap:       map[string]cron.EntryID{},
	}

	gateProxy.cron.Start()

	gateProxy.PacketConn = pktconn.NewPacketConn(context2.Background(), _conn, gateProxy)
	return gateProxy
}

func (gp *GateProxy) String() string {
	return fmt.Sprintf("GateProxy<gate:%d channel:%d addr:%s>", gp.gateID, gp.dispatcherChannelID, gp.RemoteAddr())
}

func (gp *GateProxy) serve() {
	defer func() {
		gp.CloseAll()
		if err := recover(); err != nil {
			log.Errorf("%s error: %s", gp, err.(error))
		} else {
			log.Debugf("%s disconnected", gp)
		}
	}()

	err := gp.ReceiveChan(gameServer.gatePacketQueue)
	if err != nil {
		log.Panic(err)
	}
}

func (gp *GateProxy) CloseAll() {
	defer func() {
		gp.cron.Stop()
		gameServer.onClientProxyClose(gp)
	}()

	err := gp.Close()
	if err != nil {
		log.Errorf("关闭客户端连接失败:%s", err)
	}
}

// ============================================================================业务处理
func (gp *GateProxy) handleGameLogic(pkt *pktconn.Packet) {
	gameReq := &proto.GameReq{}
	pkt.ReadData(gameReq)
	entityID := pkt.ReadInt64()
	gameProcess(gp, entityID, gameReq)
}

func (gp *GateProxy) handle3002(pkt *pktconn.Packet) {
	req := &proto.GameDispatcherChannelInfoReq{}
	pkt.ReadData(req)
	if context.GetOneConfig().Nacos.Instance.Service != req.Game {
		gp.SendGateMsg(proto.GameDispatcherChannelInfoFromDispatcherAck, &proto.GameDispatcherChannelInfoResp{
			Success: false,
			Msg:     "Game not match",
		})
		log.Error("%s Game not match", gp)

		gp.CloseAll()
	}

	gameServer.gpMutex.RLock()

	for _, proxy := range gameServer.gateProxies {
		if proxy.gateID == req.GateID && proxy.dispatcherChannelID == req.ChannelID {
			gp.CloseAll()
			break
		}
	}
	gameServer.gpMutex.RUnlock()

	gp.gateID = req.GateID
	gp.dispatcherChannelID = req.ChannelID

	gameServer.addGateProxy(gp)

	gp.SendGateMsg(proto.GameDispatcherChannelInfoFromDispatcherAck, &proto.GameDispatcherChannelInfoResp{
		Success: true,
	})
}

func (gp *GateProxy) handle3003(pkt *pktconn.Packet) {
	req := &proto.NewPlayerConnectionReq{}
	pkt.ReadData(req)

	player := GetPlayer(req.EntityID)
	if player == nil {
		player = AddPlayer(req.EntityID, gp.gateID)
		player.UpdateStatus(PlayerStatusOnline)
		gameServer.JoinScene(SceneTypeLobby, player)
	} else {
		// TODO
		log.Warnf("player:<%d> already exists", req.EntityID)
	}

}

func (gp *GateProxy) handle3004(pkt *pktconn.Packet) {
	req := &proto.PlayerDisconnectedReq{}
	pkt.ReadData(req)

	RemovePlayer(req.EntityID)
}

// ============================================================================基础协议

func (gp *GateProxy) SendGateMsg(msgType uint16, msg interface{}) {
	packet := pktconn.NewPacket()
	packet.WriteUint16(msgType)

	if msg != nil {
		packet.AppendData(msg)
	}

	gp.SendAndRelease(packet)
}

func (gp *GateProxy) SendHeartBeatAck() {
	gp.SendGateMsg(proto.HeartbeatFromDispatcherAck, nil)
}

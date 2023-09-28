package proxy

import (
	context2 "context"
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/common_proto"
	"go-one/common/consts"
	"go-one/common/context"
	"go-one/common/log"
	"go-one/common/pktconn"
	"go-one/game/common"
	"go-one/game/player"
	"go-one/game/scene_center"
	"net"
	"time"
)

type IGameServer interface {
	OnClientProxyClose(gp *GateProxy)
	AddGateProxy(gp *GateProxy)
}

type IGameProcessor interface {
	Process(gp *GateProxy, entityID int64, req *common_proto.GameReq)
}

// GateProxy is a Game client connections managed by gate
type GateProxy struct {
	*pktconn.PacketConn
	gameServer          IGameServer
	gameProcessor       IGameProcessor
	ProxyID             string
	DispatcherChannelID uint8
	GateID              uint8

	HeartbeatTime time.Time
	Cron          *cron.Cron
	cronMap       map[string]cron.EntryID
}

func NewClientProxy(_conn net.Conn, gameServer IGameServer, gameProcessor IGameProcessor) *GateProxy {
	_conn = pktconn.NewBufferedConn(_conn, consts.BufferedReadBufferSize, consts.BufferedWriteBufferSize)
	gateProxy := &GateProxy{
		gameServer:    gameServer,
		gameProcessor: gameProcessor,
		ProxyID:       context.NextClientID(),
		HeartbeatTime: time.Now(),
		Cron:          cron.New(cron.WithSeconds()),
		cronMap:       map[string]cron.EntryID{},
	}

	gateProxy.Cron.Start()

	gateProxy.PacketConn = pktconn.NewPacketConn(context2.Background(), _conn, gateProxy)
	return gateProxy
}

func (gp *GateProxy) String() string {
	return fmt.Sprintf("GateProxy<gate:%d channel:%d addr:%s>", gp.GateID, gp.DispatcherChannelID, gp.RemoteAddr())
}

func (gp *GateProxy) Serve(gatePacketQueue chan *pktconn.Packet) {
	defer func() {
		gp.CloseAll()
		if err := recover(); err != nil {
			log.Errorf("%s error: %s", gp, err.(error))
		} else {
			log.Debugf("%s disconnected", gp)
		}
	}()

	err := gp.ReceiveChan(gatePacketQueue)
	if err != nil {
		log.Error(err)
	}
}

func (gp *GateProxy) CloseAll() {
	defer func() {
		gp.Cron.Stop()
		gp.gameServer.OnClientProxyClose(gp)
	}()

	err := gp.Close()
	if err != nil {
		log.Errorf("关闭客户端连接失败:%s", err)
	}
}

// ============================================================================业务处理

func (gp *GateProxy) HandleGameLogic(pkt *pktconn.Packet) {
	gameReq := &common_proto.GameReq{}
	pkt.ReadData(gameReq)
	entityID := pkt.ReadInt64()
	gp.gameProcessor.Process(gp, entityID, gameReq)
}

func (gp *GateProxy) Handle3002(pkt *pktconn.Packet) {
	req := &common_proto.GameDispatcherChannelInfoReq{}
	pkt.ReadData(req)
	if context.GetOneConfig().Nacos.Instance.Service != req.Game {
		gp.SendGateMsg(common_proto.GameDispatcherChannelInfoFromDispatcherAck, &common_proto.GameDispatcherChannelInfoResp{
			Success: false,
			Msg:     "Game not match",
		})
		log.Error("%s Game not match", gp)

		gp.CloseAll()
	}

	gp.GateID = req.GateID
	gp.DispatcherChannelID = req.ChannelID

	gp.gameServer.AddGateProxy(gp)

	gp.SendGateMsg(common_proto.GameDispatcherChannelInfoFromDispatcherAck, &common_proto.GameDispatcherChannelInfoResp{
		Success: true,
	})
}

func (gp *GateProxy) Handle3003(pkt *pktconn.Packet) {
	req := &common_proto.NewPlayerConnectionReq{}
	pkt.ReadData(req)

	p := player.GetPlayer(req.EntityID)
	if p == nil {
		p = player.AddPlayer(req.EntityID, gp.GateID)
		p.UpdateStatus(common.PlayerStatusOnline)
		scene_center.JoinScene(common.SceneTypeLobby, 0, p)
	} else {
		p.UpdateStatus(common.PlayerStatusOnline)
		scene_center.ReJoinScene(p)
	}

}

func (gp *GateProxy) Handle3004(pkt *pktconn.Packet) {
	req := &common_proto.PlayerDisconnectedReq{}
	pkt.ReadData(req)

	player.RemovePlayer(req.EntityID)
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
	gp.SendGateMsg(common_proto.HeartbeatFromDispatcherAck, nil)
}

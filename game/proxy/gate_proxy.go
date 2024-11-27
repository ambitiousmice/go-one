package proxy

import (
	context2 "context"
	"fmt"
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/pktconn"
	"github.com/ambitiousmice/go-one/game/common"
	"github.com/ambitiousmice/go-one/game/entity"
	"github.com/robfig/cron/v3"
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
	GateClusterID       uint8

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
	return fmt.Sprintf("GateProxy<gate:%d channel:%d addr:%s>", gp.GateClusterID, gp.DispatcherChannelID, gp.RemoteAddr())
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

	gp.GateClusterID = uint8(req.GateClusterID)
	gp.DispatcherChannelID = uint8(req.ChannelID)

	gp.gameServer.AddGateProxy(gp)

	gp.SendGateMsg(common_proto.GameDispatcherChannelInfoFromDispatcherAck, &common_proto.GameDispatcherChannelInfoResp{
		Success: true,
	})
}

func (gp *GateProxy) Handle3003(pkt *pktconn.Packet) {
	req := &common_proto.NewPlayerConnectionReq{}
	pkt.ReadData(req)

	p := entity.GetPlayer(req.EntityID)

	if p == nil {
		log.Debugf("收到新玩家连接:%d", req.EntityID)
		p = entity.AddPlayer(req.EntityID, req.Region, gp.GateClusterID)
		p.UpdateStatus(common.PlayerStatusOnline)
		p.JoinScene(common.SceneTypeLobby, 0)
	} else {
		log.Debugf("收到老玩家重连:%d", req.EntityID)
		p.UpdateStatus(common.PlayerStatusOnline)
		p.ReJoinScene()
	}

}

func (gp *GateProxy) Handle3004(pkt *pktconn.Packet) {
	req := &common_proto.PlayerDisconnectedReq{}
	pkt.ReadData(req)

	player := entity.GetPlayer(req.EntityID)
	if player != nil && player.I != nil {
		player.I.OnClientDisconnected()
	}

	entity.RemovePlayer(req.EntityID)
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

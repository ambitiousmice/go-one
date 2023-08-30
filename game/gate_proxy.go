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

// GateProxy is a game client connections managed by gate
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

func (cp *GateProxy) String() string {
	return fmt.Sprintf("GateProxy<gate:%d channel:%d addr:%s>", cp.gateID, cp.dispatcherChannelID, cp.RemoteAddr())
}

func (cp *GateProxy) serve() {
	defer func() {
		cp.CloseAll()
		if err := recover(); err != nil {
			log.Errorf("%s error: %s", cp, err.(error))
		} else {
			log.Debugf("%s disconnected", cp)
		}
	}()

	err := cp.ReceiveChan(gameServer.gatePacketQueue)
	if err != nil {
		log.Panic(err)
	}
}

func (cp *GateProxy) CloseAll() {
	defer func() {
		cp.cron.Stop()
		gameServer.onClientProxyClose(cp)
	}()

	err := cp.Close()
	if err != nil {
		log.Errorf("关闭客户端连接失败:%s", err)
	}
}

// ============================================================================业务处理

func (cp *GateProxy) handle3002(pkt *pktconn.Packet) {
	req := &proto.GameDispatcherChannelInfoReq{}
	pkt.ReadData(req)
	if context.GetOneConfig().Nacos.Instance.Service != req.Game {
		cp.SendGateMsg(proto.GameDispatcherChannelInfoFromDispatcherAck, &proto.GameDispatcherChannelInfoResp{
			Success: false,
			Msg:     "game not match",
		})
		log.Error("%s game not match", cp)

		cp.CloseAll()
	}

	for _, proxy := range gameServer.gateProxies {
		if proxy.gateID == req.GateID && proxy.dispatcherChannelID == req.ChannelID {
			cp.CloseAll()
			break
		}
	}

	cp.gateID = req.GateID
	cp.dispatcherChannelID = req.ChannelID

	cp.SendGateMsg(proto.GameDispatcherChannelInfoFromDispatcherAck, &proto.GameDispatcherChannelInfoResp{
		Success: true,
	})
}

func (cp *GateProxy) handle3003(pkt *pktconn.Packet) {
	req := &proto.NewPlayerConnectionReq{}
	pkt.ReadData(req)

	basePlayer := NewBasePlayer(req.ClientID, req.EntityID)

	basePlayer.gateProxy = cp

	AddPlayer(basePlayer)
}

func (cp *GateProxy) handle3004(pkt *pktconn.Packet) {
	req := &proto.NewPlayerConnectionReq{}
	pkt.ReadData(req)

	basePlayer := NewBasePlayer(req.ClientID, req.EntityID)

	basePlayer.gateProxy = cp

	AddPlayer(basePlayer)
}

// ============================================================================基础协议

func (cp *GateProxy) SendGateMsg(msgType uint16, msg interface{}) {
	packet := pktconn.NewPacket()
	packet.WriteUint16(msgType)

	if msg != nil {
		packet.AppendData(msg)
	}

	cp.SendAndRelease(packet)
}

func (cp *GateProxy) SendHeartBeatAck() {
	cp.SendGateMsg(proto.HeartbeatFromDispatcherAck, nil)
}

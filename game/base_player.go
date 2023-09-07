package game

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/log"
	"go-one/common/pktconn"
	"go-one/common/proto"
)

type BasePlayer struct {
	entityID int64
	gateID   uint8

	I IPlayer

	cron    *cron.Cron
	cronMap map[string]cron.EntryID
	attrMap map[string]interface{}
}

func NewBasePlayer(entityID int64, gateID uint8) *BasePlayer {
	crontab := cron.New(cron.WithSeconds())
	crontab.Start()
	return &BasePlayer{
		entityID: entityID,
		gateID:   gateID,
		cron:     crontab,
		cronMap:  map[string]cron.EntryID{},
		attrMap:  map[string]interface{}{},
	}
}

type IPlayer interface {
	OnInit()
	OnAttrsReady()
	OnCreated()
	OnDestroy()
	OnClientConnected()
	OnClientDisconnected()
}

func (p *BasePlayer) String() string {
	return fmt.Sprintf("BasePlayer:<%d> gateID:<%d>", p.entityID, p.gateID)
}

func (p *BasePlayer) SendCommonErrorMsg(error string) {
	p.SendGameMsg(&proto.GameResp{
		Cmd:  proto.Error,
		Data: []byte(error),
	})
}

func (p *BasePlayer) SendErrorMsg(cmd uint16, error string) {
	p.SendGameMsg(&proto.GameResp{
		Cmd:  cmd,
		Code: proto.Error,
		Data: []byte(error),
	})
}

func (p *BasePlayer) SendGameMsg(resp *proto.GameResp) {
	packet := pktconn.NewPacket()
	packet.WriteUint16(proto.GameMethodFromClientAck)

	if resp != nil {
		packet.AppendData(resp)
	}

	packet.WriteInt64(p.entityID)

	gateProxy := gameServer.getGateProxyByGateID(p.gateID)
	if gateProxy == nil {
		log.Errorf("%s not found gate proxy", p)
		return
	}

	err := gateProxy.SendAndRelease(packet)

	if err != nil {
		log.Errorf("%s send game msg error: %s", p, err)
	}
}

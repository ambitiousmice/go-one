package game

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/pktconn"
	"go-one/common/proto"
)

type BasePlayer struct {
	gateProxy *GateProxy
	clientID  string
	entityID  int64

	I IPlayer

	cron    *cron.Cron
	cronMap map[string]cron.EntryID
	attrMap map[string]interface{}
}

func NewBasePlayer(clientID string, entityID int64) *BasePlayer {
	crontab := cron.New(cron.WithSeconds())
	crontab.Start()
	return &BasePlayer{
		clientID: clientID,
		entityID: entityID,
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
	return fmt.Sprintf("%s<%d><%d>", p.clientID, p.entityID, p.gateProxy.gateID)
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

	packet.AppendClientID(p.clientID)

	p.gateProxy.SendAndRelease(packet)
}

package player

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/common_proto"
	"go-one/common/consts"
	"go-one/common/log"
	"go-one/common/pktconn"
	"sync"
)

type IGameServer interface {
	SendAndRelease(gateClusterID uint8, packet *pktconn.Packet)
}

var gameServer IGameServer

func SetGameServer(gs IGameServer) {
	gameServer = gs
}

type IScene interface {
	GetSceneType() string
}

type BasePlayer struct {
	sync.RWMutex
	EntityID      int64
	gateClusterID uint8
	SceneType     string
	SceneID       int64
	status        uint8

	cron    *cron.Cron
	cronMap map[string]cron.EntryID
	attrMap map[string]interface{}
}

func NewBasePlayer(entityID int64, gateClusterID uint8) *BasePlayer {
	crontab := cron.New(cron.WithSeconds())
	crontab.Start()
	return &BasePlayer{
		EntityID:      entityID,
		gateClusterID: gateClusterID,
		cron:          crontab,
		cronMap:       map[string]cron.EntryID{},
		attrMap:       map[string]interface{}{},
	}
}

func (p *BasePlayer) String() string {
	return fmt.Sprintf("player info: EntityID=<%d>, gateClusterID=<%d>", p.EntityID, p.gateClusterID)
}

func (p *BasePlayer) SendCommonErrorMsg(error string) {
	p.SendGameMsg(&common_proto.GameResp{
		Cmd:  common_proto.Error,
		Data: []byte(error),
	})
}

func (p *BasePlayer) SendErrorMsg(cmd uint16, error string) {
	p.SendGameMsg(&common_proto.GameResp{
		Cmd:  cmd,
		Code: consts.ErrorCommon,
		Data: []byte(error),
	})
}

func (p *BasePlayer) SendGameMsg(resp *common_proto.GameResp) {
	packet := pktconn.NewPacket()
	packet.WriteUint16(common_proto.GameMethodFromClientAck)

	if resp != nil {
		packet.AppendData(resp)
	}

	packet.WriteInt64(p.EntityID)

	gameServer.SendAndRelease(p.gateClusterID, packet)
}

func (p *BasePlayer) SendGameData(cmd uint16, data interface{}) {
	packet := pktconn.NewPacket()
	packet.WriteUint16(common_proto.GameMethodFromClientAck)

	dataByte, err := pktconn.MSG_PACKER.PackMsg(data, nil)
	if err != nil {
		log.Errorf("%s pack msg error: %s", p, err)
		return
	}

	resp := &common_proto.GameResp{
		Cmd:  cmd,
		Data: dataByte,
	}

	if resp != nil {
		packet.AppendData(resp)
	}

	packet.WriteInt64(p.EntityID)

	gameServer.SendAndRelease(p.gateClusterID, packet)
}

func (p *BasePlayer) UpdateStatus(status uint8) {
	p.Lock()
	defer p.Unlock()
	p.status = status
}

package entity

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/common_proto"
	"go-one/common/consts"
	"go-one/common/log"
	"go-one/common/pktconn"
	"go-one/game/aoi"
	"go-one/game/common"
	"sync"
)

type IGameServer interface {
	SendAndRelease(gateClusterID uint8, packet *pktconn.Packet)
}

var gameServer IGameServer

func SetGameServer(gs IGameServer) {
	gameServer = gs
}

type BasePlayer struct {
	sync.RWMutex
	EntityID      int64
	Region        int32
	gateClusterID uint8
	Scene         *Scene
	status        uint8

	aoiMutex     sync.RWMutex
	Position     Vector3
	InterestedIn BasePlayerSet
	InterestedBy BasePlayerSet
	AOI          aoi.AOI
	Yaw          common.Yaw
	Speed        common.Speed
	SyncAOI      bool

	cron    *cron.Cron
	cronMap map[string]cron.EntryID
	attrMap map[string]interface{}
}

func NewBasePlayer(entityID int64, region int32, gateClusterID uint8) *BasePlayer {
	crontab := cron.New(cron.WithSeconds())
	crontab.Start()
	basePlayer := &BasePlayer{
		EntityID:      entityID,
		Region:        region,
		gateClusterID: gateClusterID,
		cron:          crontab,
		cronMap:       map[string]cron.EntryID{},
		attrMap:       map[string]interface{}{},
		InterestedIn:  make(BasePlayerSet),
		InterestedBy:  make(BasePlayerSet),
	}

	aoi.InitAOI(&basePlayer.AOI, 0, basePlayer, basePlayer)

	return basePlayer
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
		Cmd:  int32(cmd),
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
		Cmd:  int32(cmd),
		Data: dataByte,
	}

	if resp != nil {
		packet.AppendData(resp)
	}

	packet.WriteInt64(p.EntityID)

	gameServer.SendAndRelease(p.gateClusterID, packet)
}

func (p *BasePlayer) SendCreateEntity(createPlayer *BasePlayer) {

	//var clientData map[string]interface{}

	createPlayerData := &common_proto.OnCreateEntity{
		EntityID: createPlayer.EntityID,
		X:        float32(createPlayer.Position.X),
		Y:        float32(createPlayer.Position.Y),
		Z:        float32(createPlayer.Position.Z),
		Yaw:      float32(createPlayer.Yaw),
		Speed:    float32(createPlayer.Speed),
	}

	p.SendGameData(common_proto.CreateEntity, createPlayerData)
}

func (p *BasePlayer) SendDestroyEntity(player *BasePlayer) {

	destroyPlayerData := &common_proto.OnDestroyEntity{
		EntityID: player.EntityID,
	}

	p.SendGameData(common_proto.DestroyEntity, destroyPlayerData)
}

func (p *BasePlayer) UpdateStatus(status uint8) {
	p.Lock()
	defer p.Unlock()
	p.status = status
}

func (p *BasePlayer) OnEnterAOI(otherAoi *aoi.AOI) {
	p.interest(otherAoi.Data.(*BasePlayer))
}

func (p *BasePlayer) OnLeaveAOI(otherAoi *aoi.AOI) {
	p.uninterested(otherAoi.Data.(*BasePlayer))
}

func (p *BasePlayer) interest(other *BasePlayer) {
	p.aoiMutex.Lock()
	p.InterestedIn.Add(other)
	p.aoiMutex.Unlock()

	other.aoiMutex.Lock()
	other.InterestedBy.Add(p)
	other.aoiMutex.Unlock()
	p.SendCreateEntity(other)
}

func (p *BasePlayer) uninterested(other *BasePlayer) {
	p.aoiMutex.Lock()
	p.InterestedIn.Del(other)
	p.aoiMutex.Unlock()

	other.aoiMutex.Lock()
	other.InterestedBy.Del(p)
	other.aoiMutex.Unlock()
	p.SendDestroyEntity(other)
}

func (p *BasePlayer) IsInterestedIn(other *BasePlayer) bool {
	return p.InterestedIn.Contains(other)
}

func (p *BasePlayer) DistanceTo(other *BasePlayer) common.Coord {
	return p.Position.DistanceTo(other.Position)
}

func (p *BasePlayer) CollectAOISyncInfos() {
	syncInfoLength := len(p.InterestedBy)
	if syncInfoLength == 0 {
		return
	}

	syncInfos := make([]*common_proto.AOISyncInfo, 0)
	p.aoiMutex.RLock()
	for neighbor := range p.InterestedBy {
		aoiSyncInfo := &common_proto.AOISyncInfo{
			EntityID: neighbor.EntityID,
			X:        float32(neighbor.Position.X),
			Y:        float32(neighbor.Position.Y),
			Z:        float32(neighbor.Position.Z),
			Yaw:      float32(neighbor.Yaw),
			Speed:    float32(neighbor.Speed),
		}
		syncInfos = append(syncInfos, aoiSyncInfo)
	}
	p.aoiMutex.RUnlock()

	syncInfoBatch := &common_proto.AOISyncInfoBatch{
		SyncInfos: syncInfos,
	}

	p.SendGameData(common_proto.AOISync, syncInfoBatch)
}

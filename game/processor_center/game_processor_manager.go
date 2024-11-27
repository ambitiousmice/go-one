package processor_center

import (
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/pool/fixed_channel_pool"
	"github.com/ambitiousmice/go-one/game/entity"
	"github.com/ambitiousmice/go-one/game/processor_center/base_processor"
	"github.com/ambitiousmice/go-one/game/proxy"
	"strconv"
)

var GPM = &GameProcessManager{
	processContext: map[uint16]Processor{},
}

func init() {
	GPM.RegisterProcessor(&base_processor.JoinSceneProcessor{})
	GPM.RegisterProcessor(&base_processor.LeaveSceneProcessor{})
	GPM.RegisterProcessor(&base_processor.MoveProcessor{})
}

type GameProcessManager struct {
	processContext map[uint16]Processor
}

func (gpm *GameProcessManager) RegisterProcessor(p Processor) {
	processor := gpm.processContext[p.GetCmd()]
	if processor != nil {
		log.Panic("duplicate processor_center: " + strconv.Itoa(int(p.GetCmd())))
	}
	gpm.processContext[p.GetCmd()] = p
}

func (gpm *GameProcessManager) Process(gp *proxy.GateProxy, entityID int64, req *common_proto.GameReq) {
	p := entity.GetPlayer(entityID)
	if p == nil {
		log.Warnf("p:<%d> not found", entityID)
		/*p = game.AddPosition(entityID, gp.gateClusterID)
		p.UpdateStatus(game.PlayerStatusOnline)*/
	}
	processor := gpm.processContext[uint16(req.Cmd)]
	if processor == nil {
		log.Errorf("player:%d send invalid cmd: %d", entityID, req.Cmd)
		p.SendCommonErrorMsg("invalid cmd")
		return
	}
	fixed_channel_pool.Submit(entityID, func() {
		log.Infof("player:%d send cmd: %d", entityID, req.Cmd)
		processor.Process(p, req.Param)
	})
}

package base_processor

import (
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/game/common"
	"github.com/ambitiousmice/go-one/game/entity"
)

type MoveProcessor struct {
}

func (t *MoveProcessor) Process(player *entity.Player, param []byte) {
	moveReq := &common_proto.MoveReq{}
	err := common.UnPackMsg(param, moveReq)
	if err != nil {
		player.SendCommonErrorMsg(consts.ParamError)
		return
	}
	player.Move(moveReq)
	//log.Infof("player %d move:%v", player.EntityID, moveReq)
}

func (t *MoveProcessor) GetCmd() uint16 {
	return common_proto.Move
}

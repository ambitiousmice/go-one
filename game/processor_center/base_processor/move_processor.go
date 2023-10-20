package base_processor

import (
	"go-one/common/common_proto"
	"go-one/common/consts"
	"go-one/game/common"
	"go-one/game/entity"
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

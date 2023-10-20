package base_processor

import (
	"go-one/common/common_proto"
	"go-one/common/consts"
	"go-one/game/common"
	"go-one/game/entity"
)

type JoinSceneProcessor struct {
}

func (p *JoinSceneProcessor) Process(player *entity.Player, param []byte) {
	joinSceneReq := &common_proto.JoinSceneReq{}
	err := common.UnPackMsg(param, joinSceneReq)
	if err != nil {
		player.SendCommonErrorMsg(consts.ParamError)
		return
	}

	player.JoinScene(joinSceneReq.SceneType, joinSceneReq.SceneID)
}

func (p *JoinSceneProcessor) GetCmd() uint16 {
	return common_proto.JoinScene
}

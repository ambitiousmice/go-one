package base_processor

import (
	"go-one/common/common_proto"
	"go-one/game/player"
	"go-one/game/scene_center"
)

type JoinSceneProcessor struct {
}

func (p *JoinSceneProcessor) Process(player *player.Player, param []byte) {
	joinSceneReq := &common_proto.JoinSceneReq{}
	UnPackMsg(player, param, joinSceneReq)

	sceneType := player.SceneType
	if sceneType != "" {
		if sceneType == joinSceneReq.SceneType {
			scene_center.ReJoinScene(player)
			return
		}

		scene_center.Leave(player)
	}

	scene_center.JoinScene(joinSceneReq.SceneType, joinSceneReq.SceneID, player)
}

func (p *JoinSceneProcessor) GetCmd() uint16 {
	return common_proto.JoinScene
}

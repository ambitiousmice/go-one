package processor

import (
	"go-one/common/proto"
	"go-one/game"
)

type JoinSceneProcessor struct {
}

func (p *JoinSceneProcessor) Process(player *game.Player, param []byte) {
	var joinSceneReq proto.JoinSceneReq
	UnPackMsg(player, param, joinSceneReq)

	scene := player.Scene
	if scene != nil {
		if scene.Type == joinSceneReq.SceneType {
			scene.Join(player)
			return
		}

		scene.Leave(player)
	}

	sceneManager := game.GetGameServer().GetSceneManager(joinSceneReq.SceneType)

	if joinSceneReq.SceneID == 0 {
		scene = sceneManager.GetSceneByStrategy()
	} else {
		scene = sceneManager.GetScene(joinSceneReq.SceneID)
	}

	if scene == nil {
		player.SendCommonErrorMsg(game.ServerIsFull)
	}

	scene.Join(player)
}

func (p *JoinSceneProcessor) GetCmd() uint16 {
	return proto.JoinScene
}

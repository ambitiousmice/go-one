package proccessor

import (
	"go-one/common/consts"
	"go-one/common/pktconn"
	"go-one/common/proto"
	"go-one/game"
)

type JoinSceneProcessor struct {
}

func (p *JoinSceneProcessor) Process(player *game.Player, param []byte) {
	joinSceneReq := &proto.JoinSceneReq{}
	err := pktconn.MSG_PACKER.UnpackMsg(param, joinSceneReq)
	if err != nil {
		player.SendCommonErrorMsg(consts.ParamError)
	}

	room := player.Scene
	if room != nil {
		room.Leave(player)
	}

	sceneManager := game.GetGameServer().GetSceneManager(joinSceneReq.SceneType)

	if joinSceneReq.SceneID == 0 {
		room = sceneManager.GetSceneByStrategy()
	} else {
		room = sceneManager.GetScene(joinSceneReq.SceneID)
	}

	if room == nil {
		player.SendCommonErrorMsg(game.ServerIsFull)
	}

	room.Join(player)
}

func (p *JoinSceneProcessor) GetCmd() uint16 {
	return proto.JoinScene
}

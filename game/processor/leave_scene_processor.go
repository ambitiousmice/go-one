package processor

import (
	"go-one/common/consts"
	"go-one/common/pktconn"
	"go-one/common/proto"
	"go-one/game"
)

type LeaveSceneProcessor struct {
}

func (p *LeaveSceneProcessor) Process(player *game.Player, param []byte) {
	leaveSceneReq := &proto.LeaveSceneReq{}
	err := pktconn.MSG_PACKER.UnpackMsg(param, leaveSceneReq)
	if err != nil {
		player.SendCommonErrorMsg(consts.ParamError)
	}

	sceneManager := game.GetGameServer().GetSceneManager(leaveSceneReq.SceneType)
	var room *game.Scene
	if leaveSceneReq.SceneID == 0 {
		room = sceneManager.GetSceneByStrategy()
	} else {
		room = sceneManager.GetScene(leaveSceneReq.SceneID)
	}

	if room == nil {
		player.SendCommonErrorMsg(game.ServerIsFull)
	}
	room.Leave(player)
}

func (p *LeaveSceneProcessor) GetCmd() uint16 {
	return proto.JoinScene
}

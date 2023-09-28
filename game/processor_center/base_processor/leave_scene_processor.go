package base_processor

import (
	"go-one/common/common_proto"
	"go-one/common/consts"
	"go-one/common/pktconn"
	"go-one/game/common"
	"go-one/game/player"
	"go-one/game/scene_center"
)

type LeaveSceneProcessor struct {
}

func (p *LeaveSceneProcessor) Process(player *player.Player, param []byte) {
	leaveSceneReq := &common_proto.LeaveSceneReq{}
	err := pktconn.MSG_PACKER.UnpackMsg(param, leaveSceneReq)
	if err != nil {
		player.SendCommonErrorMsg(consts.ParamError)
	}

	sceneManager := scene_center.GetSceneManager(leaveSceneReq.SceneType)
	var room *scene_center.Scene
	if leaveSceneReq.SceneID == 0 {
		room = sceneManager.GetSceneByStrategy()
	} else {
		room = sceneManager.GetScene(leaveSceneReq.SceneID)
	}

	if room == nil {
		player.SendCommonErrorMsg(common.ServerIsFull)
	}
	scene_center.Leave(player)
}

func (p *LeaveSceneProcessor) GetCmd() uint16 {
	return common_proto.LeaveScene
}

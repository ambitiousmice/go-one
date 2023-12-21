package base_processor

import (
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/pktconn"
	"github.com/ambitiousmice/go-one/game/common"
	"github.com/ambitiousmice/go-one/game/entity"
)

type LeaveSceneProcessor struct {
}

func (p *LeaveSceneProcessor) Process(player *entity.Player, param []byte) {
	leaveSceneReq := &common_proto.LeaveSceneReq{}
	err := pktconn.MSG_PACKER.UnpackMsg(param, leaveSceneReq)
	if err != nil {
		player.SendCommonErrorMsg(consts.ParamError)
		return
	}

	sceneManager := entity.GetSceneManager(leaveSceneReq.SceneType)
	var room *entity.Scene
	if leaveSceneReq.SceneID == 0 {
		room = sceneManager.GetSceneByStrategy()
	} else {
		room = sceneManager.GetScene(leaveSceneReq.SceneID)
	}

	if room == nil {
		player.SendCommonErrorMsg(common.ServerIsFull)
	}
	player.LeaveScene()
}

func (p *LeaveSceneProcessor) GetCmd() uint16 {
	return common_proto.LeaveScene
}

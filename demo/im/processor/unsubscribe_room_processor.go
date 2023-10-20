package processor

import (
	"go-one/common/consts"
	"go-one/demo/im/proto"
	scene2 "go-one/demo/im/scene"
	"go-one/game/common"
	"go-one/game/entity"
)

type UnsubscribeRoomProcessor struct {
}

func (t *UnsubscribeRoomProcessor) Process(player *entity.Player, param []byte) {
	subscribeRoomReq := proto.SubscribeRoomReq{}
	err := common.UnPackMsg(param, subscribeRoomReq)
	if err != nil {
		player.SendCommonErrorMsg(consts.ParamError)
		return
	}

	scene := player.Scene
	if scene == nil {
		return
	}

	scene.I.(*scene2.ChatScene).RoomManager.UnsubscribeRoom(player, subscribeRoomReq.RoomID)
}

func (t *UnsubscribeRoomProcessor) GetCmd() uint16 {
	return proto.UnsubscribeRoomAck
}

package processor

import (
	"go-one/demo/im/proto"
	scene2 "go-one/demo/im/scene"
	"go-one/game/player"
	"go-one/game/scene_center"
)

type UnsubscribeRoomProcessor struct {
}

func (t *UnsubscribeRoomProcessor) Process(player *player.Player, param []byte) {
	subscribeRoomReq := proto.SubscribeRoomReq{}
	UnPackMsg(player, param, subscribeRoomReq)

	scene := scene_center.GetSceneByPlayer(player)
	scene.I.(*scene2.ChatScene).RoomManager.UnsubscribeRoom(player, subscribeRoomReq.RoomID)
}

func (t *UnsubscribeRoomProcessor) GetCmd() uint16 {
	return proto.UnsubscribeRoomAck
}

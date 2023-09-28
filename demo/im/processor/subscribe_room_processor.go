package processor

import (
	"go-one/demo/im/proto"
	scene2 "go-one/demo/im/scene"
	"go-one/game/player"
	"go-one/game/scene_center"
)

type SubscribeRoomProcessor struct {
}

func (t *SubscribeRoomProcessor) Process(player *player.Player, param []byte) {
	subscribeRoomReq := &proto.SubscribeRoomReq{}
	UnPackMsg(player, param, subscribeRoomReq)

	scene := scene_center.GetSceneByPlayer(player)
	scene.I.(*scene2.ChatScene).RoomManager.SubscribeRoom(player, subscribeRoomReq.RoomID)

}

func (t *SubscribeRoomProcessor) GetCmd() uint16 {
	return proto.SubscribeRoom
}

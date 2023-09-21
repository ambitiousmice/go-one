package processor

import (
	"go-one/demo/im/proto"
	"go-one/demo/im/room"
	"go-one/game"
)

type SubscribeRoomProcessor struct {
}

func (t *SubscribeRoomProcessor) Process(player *game.Player, param []byte) {
	subscribeRoomReq := &proto.SubscribeRoomReq{}
	UnPackMsg(player, param, subscribeRoomReq)

	room.CRM.SubscribeRoom(player, subscribeRoomReq.RoomID)
}

func (t *SubscribeRoomProcessor) GetCmd() uint16 {
	return proto.SubscribeRoom
}

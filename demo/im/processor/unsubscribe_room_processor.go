package processor

import (
	"go-one/demo/im/proto"
	"go-one/demo/im/room"
	"go-one/game"
)

type UnsubscribeRoomProcessor struct {
}

func (t *UnsubscribeRoomProcessor) Process(player *game.Player, param []byte) {
	subscribeRoomReq := proto.SubscribeRoomReq{}
	UnPackMsg(player, param, subscribeRoomReq)

	room.CRM.UnsubscribeRoom(player, subscribeRoomReq.RoomID)
}

func (t *UnsubscribeRoomProcessor) GetCmd() uint16 {
	return proto.UnsubscribeRoomAck
}

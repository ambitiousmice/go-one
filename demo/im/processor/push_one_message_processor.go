package processor

import (
	"go-one/demo/im/proto"
	"go-one/demo/im/room"
	"go-one/game"
)

type PushOneMessageProcessor struct {
}

func (t *PushOneMessageProcessor) Process(player *game.Player, param []byte) {
	pushMessageReq := &proto.PushMessageReq{}
	UnPackMsg(player, param, pushMessageReq)

	room.CRM.SubscribeRoom(player, pushMessageReq.RoomID)
}

func (t *PushOneMessageProcessor) GetCmd() uint16 {
	return proto.PushOneMessage
}

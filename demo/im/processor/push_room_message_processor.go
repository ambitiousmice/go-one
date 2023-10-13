package processor

import (
	"go-one/common/mq/kafka"
	"go-one/common/utils"
	"go-one/demo/im/message_center"
	"go-one/demo/im/proto"
	"go-one/game/player"
)

type PushRoomMessageProcessor struct {
}

func (t *PushRoomMessageProcessor) Process(player *player.Player, param []byte) {
	pushMessageReq := &proto.PushMessageReq{}
	UnPackMsg(player, param, pushMessageReq)

	pushMessageReq.From = player.EntityID

	//message_center.RoomMessageHandler(pushMessageReq)
	kafka.Producer.SendMessage(message_center.Room, utils.ToString(pushMessageReq.RoomID), pushMessageReq)

}

func (t *PushRoomMessageProcessor) GetCmd() uint16 {
	return proto.PushRoomMessage
}

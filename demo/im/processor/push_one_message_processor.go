package processor

import (
	"go-one/demo/im/message_center"
	"go-one/demo/im/proto"
	"go-one/game/player"
)

type PushOneMessageProcessor struct {
}

func (t *PushOneMessageProcessor) Process(player *player.Player, param []byte) {
	pushMessageReq := &proto.PushMessageReq{}
	UnPackMsg(player, param, pushMessageReq)

	pushMessageReq.From = player.EntityID

	message_center.OneMessageHandler(pushMessageReq)
	//kafka.Producer.SendMessage(message_center.One, utils.ToString(pushMessageReq.To), pushMessageReq)

}

func (t *PushOneMessageProcessor) GetCmd() uint16 {
	return proto.PushOneMessage
}

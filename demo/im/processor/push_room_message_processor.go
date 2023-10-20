package processor

import (
	"go-one/common/consts"
	"go-one/common/mq/kafka"
	"go-one/common/utils"
	"go-one/demo/im/message_center"
	"go-one/demo/im/proto"
	"go-one/game/common"
	"go-one/game/entity"
)

type PushRoomMessageProcessor struct {
}

func (t *PushRoomMessageProcessor) Process(player *entity.Player, param []byte) {
	pushMessageReq := &proto.PushMessageReq{}
	err := common.UnPackMsg(param, pushMessageReq)
	if err != nil {
		player.SendCommonErrorMsg(consts.ParamError)
		return
	}
	pushMessageReq.From = player.EntityID

	//message_center.RoomMessageHandler(pushMessageReq)
	kafka.Producer.SendMessage(message_center.Room, utils.ToString(pushMessageReq.RoomID), pushMessageReq)

}

func (t *PushRoomMessageProcessor) GetCmd() uint16 {
	return proto.PushRoomMessage
}

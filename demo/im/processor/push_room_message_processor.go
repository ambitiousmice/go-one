package processor

import (
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/mq/kafka"
	"github.com/ambitiousmice/go-one/common/utils"
	"github.com/ambitiousmice/go-one/demo/im/message_center"
	"github.com/ambitiousmice/go-one/demo/im/proto"
	"github.com/ambitiousmice/go-one/game/common"
	"github.com/ambitiousmice/go-one/game/entity"
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

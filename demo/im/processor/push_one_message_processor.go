package processor

import (
	"go-one/common/consts"
	"go-one/demo/im/message_center"
	"go-one/demo/im/proto"
	"go-one/game/common"
	"go-one/game/entity"
)

type PushOneMessageProcessor struct {
}

func (t *PushOneMessageProcessor) Process(player *entity.Player, param []byte) {
	pushMessageReq := &proto.PushMessageReq{}
	err := common.UnPackMsg(param, pushMessageReq)
	if err != nil {
		player.SendCommonErrorMsg(consts.ParamError)
		return
	}

	pushMessageReq.From = player.EntityID

	message_center.OneMessageHandler(pushMessageReq)
	//kafka.Producer.SendMessage(message_center.One, utils.ToString(pushMessageReq.To), pushMessageReq)

}

func (t *PushOneMessageProcessor) GetCmd() uint16 {
	return proto.PushOneMessage
}

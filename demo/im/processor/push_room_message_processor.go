package processor

import (
	"go-one/demo/im/proto"
	"go-one/game"
)

type PushRoomMessageProcessor struct {
}

func (t *PushRoomMessageProcessor) Process(player *game.Player, param []byte) {
	pushMessageReq := &proto.PushMessageReq{}
	UnPackMsg(player, param, pushMessageReq)

}

func (t *PushRoomMessageProcessor) GetCmd() uint16 {
	return proto.PushRoomMessage
}

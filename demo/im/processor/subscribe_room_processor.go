package processor

import (
	"go-one/demo/im/proto"
	"go-one/game"
)

type SubscribeRoomProcessor struct {
}

func (t *SubscribeRoomProcessor) Process(player *game.Player, param []byte) {
	var subscribeRoomReq proto.SubscribeRoomReq
	UnPackMsg(player, param, subscribeRoomReq)

}

func (t *SubscribeRoomProcessor) GetCmd() uint16 {
	return proto.SubscribeRoom
}

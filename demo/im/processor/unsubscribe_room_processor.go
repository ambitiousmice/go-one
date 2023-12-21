package processor

import (
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/demo/im/proto"
	scene2 "github.com/ambitiousmice/go-one/demo/im/scene"
	"github.com/ambitiousmice/go-one/game/common"
	"github.com/ambitiousmice/go-one/game/entity"
)

type UnsubscribeRoomProcessor struct {
}

func (t *UnsubscribeRoomProcessor) Process(player *entity.Player, param []byte) {
	unSubscribeRoomReq := &proto.UnsubscribeRoomReq{}
	err := common.UnPackMsg(param, unSubscribeRoomReq)
	if err != nil {
		player.SendCommonErrorMsg(consts.ParamError)
		return
	}

	scene := player.Scene
	if scene == nil {
		return
	}

	scene.I.(*scene2.ChatScene).RoomManager.UnsubscribeRoom(player, unSubscribeRoomReq.RoomID)
}

func (t *UnsubscribeRoomProcessor) GetCmd() uint16 {
	return proto.UnsubscribeRoomAck
}

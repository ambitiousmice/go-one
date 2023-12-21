package processor

import (
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/demo/im/proto"
	scene2 "github.com/ambitiousmice/go-one/demo/im/scene"
	"github.com/ambitiousmice/go-one/game/common"
	"github.com/ambitiousmice/go-one/game/entity"
)

type SubscribeRoomProcessor struct {
}

func (t *SubscribeRoomProcessor) Process(player *entity.Player, param []byte) {
	subscribeRoomReq := &proto.SubscribeRoomReq{}
	err := common.UnPackMsg(param, subscribeRoomReq)
	if err != nil {
		player.SendCommonErrorMsg(consts.ParamError)
		return
	}
	scene := player.Scene
	if scene == nil {
		player.SendCommonErrorMsg("加入房间失败")
		return
	}

	scene.I.(*scene2.ChatScene).RoomManager.SubscribeRoom(player, subscribeRoomReq.RoomID)

}

func (t *SubscribeRoomProcessor) GetCmd() uint16 {
	return proto.SubscribeRoom
}

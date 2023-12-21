package message_center

import (
	"github.com/ambitiousmice/go-one/demo/im/common"
	"github.com/ambitiousmice/go-one/demo/im/proto"
	scene2 "github.com/ambitiousmice/go-one/demo/im/scene"
	"github.com/ambitiousmice/go-one/game/entity"
)

func RoomMessageHandler(msg *proto.PushMessageReq) {
	for _, scene := range entity.SceneManagerContext[common.SceneTypeChat].GetScenes() {
		room := scene.I.(*scene2.ChatScene).RoomManager.GetRoom(msg.RoomID)
		if room == nil {
			continue
		}

		room.Broadcast(&proto.ChatMessage{
			RoomID: msg.RoomID,
			From:   msg.From,
			Msg:    msg.Msg,
		})
	}
}

func OneMessageHandler(msg *proto.PushMessageReq) {
	for _, scene := range entity.SceneManagerContext[common.SceneTypeChat].GetScenes() {
		scene.PushOne(msg.To, proto.MessageAck, &proto.ChatMessage{
			RoomID: common.OneRoomID,
			From:   msg.From,
			Msg:    msg.Msg,
		})
	}
}

func BroadcastMessageHandler(msg *proto.PushMessageReq) {
	for _, scene := range entity.SceneManagerContext[common.SceneTypeChat].GetScenes() {
		scene.Broadcast(proto.MessageAck, &proto.ChatMessage{
			RoomID: common.BroadcastRoomID,
			From:   msg.From,
			Msg:    msg.Msg,
		})

	}
}

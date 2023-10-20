package message_center

import (
	"go-one/demo/im/common"
	"go-one/demo/im/proto"
	scene2 "go-one/demo/im/scene"
	"go-one/game/entity"
)

func RoomMessageHandler(msg *proto.PushMessageReq) {
	for _, scene := range entity.ManagerContext[common.SceneTypeChat].GetScenes() {
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
	for _, scene := range entity.ManagerContext[common.SceneTypeChat].GetScenes() {
		scene.PushOne(msg.To, proto.MessageAck, &proto.ChatMessage{
			RoomID: common.OneRoomID,
			From:   msg.From,
			Msg:    msg.Msg,
		})
	}
}

func BroadcastMessageHandler(msg *proto.PushMessageReq) {
	for _, scene := range entity.ManagerContext[common.SceneTypeChat].GetScenes() {
		scene.Broadcast(proto.MessageAck, &proto.ChatMessage{
			RoomID: common.BroadcastRoomID,
			From:   msg.From,
			Msg:    msg.Msg,
		})

	}
}

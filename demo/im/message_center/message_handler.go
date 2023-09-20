package message_center

import (
	"go-one/demo/im/common"
	"go-one/demo/im/proto"
	scene2 "go-one/demo/im/scene"
	"go-one/game"
)

type ChatMessage struct {
	RoomID int64
	From   int64
	To     int64
	Msg    string
}

func RoomMessageHandler(msg *ChatMessage) {
	for _, scene := range game.GetGameServer().SceneManagers[common.SceneTypeChat].GetScenes() {
		room := scene.I.(*scene2.ChatScene).RoomManager.Rooms[msg.RoomID]
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

func OneMessageHandler(msg *ChatMessage) {
	for _, scene := range game.GetGameServer().SceneManagers[common.SceneTypeChat].GetScenes() {
		toPlayer := scene.GetPlayer(msg.To)
		if toPlayer == nil {
			continue
		}

		toPlayer.SendGameData(proto.MessageAck, &proto.ChatMessage{
			RoomID: common.OneRoomID,
			From:   msg.From,
			Msg:    msg.Msg,
		})
	}
}

func BroadcastMessageHandler(msg *ChatMessage) {
	for _, scene := range game.GetGameServer().SceneManagers[common.SceneTypeChat].GetScenes() {
		scene.Broadcast(proto.MessageAck, &proto.ChatMessage{
			RoomID: common.BroadcastRoomID,
			From:   msg.From,
			Msg:    msg.Msg,
		})

	}
}

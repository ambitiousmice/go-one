package room

import "go-one/game"

var CRM *ChatRoomManager

type ChatRoomManager struct {
	subscribers map[int64]*map[int64]*game.Player
}

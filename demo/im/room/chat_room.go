package room

import (
	"go-one/demo/im/proto"
	"go-one/game"
	"sync"
)

type ChatRoom struct {
	ID   int64
	name string

	pMutex  sync.RWMutex
	players map[int64]*game.Player
}

func NewChatRoom(id int64, name string) *ChatRoom {
	return &ChatRoom{
		ID:      id,
		name:    name,
		players: make(map[int64]*game.Player),
	}
}

func (r *ChatRoom) JoinPlayer(player *game.Player) {
	r.pMutex.Lock()
	defer r.pMutex.Unlock()

	r.players[player.EntityID] = player
}

func (r *ChatRoom) Broadcast(msg *proto.ChatMessage) {
	r.pMutex.RLock()
	defer r.pMutex.RUnlock()

	for _, player := range r.players {
		player.SendGameData(proto.MessageAck, msg)
	}

}

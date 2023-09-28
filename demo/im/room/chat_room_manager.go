package room

import (
	"go-one/demo/im/proto"
	"go-one/game/player"
	"sync"
)

type ChatRoomManager struct {
	rMutex sync.RWMutex
	Rooms  map[int64]*ChatRoom
}

func NewChatRoomManager() *ChatRoomManager {
	return &ChatRoomManager{
		Rooms: make(map[int64]*ChatRoom),
	}
}

func (crm *ChatRoomManager) GetRoom(roomID int64) *ChatRoom {
	crm.rMutex.RLock()
	room := crm.Rooms[roomID]
	crm.rMutex.RUnlock()
	return room
}

func (crm *ChatRoomManager) GetRoomNotNil(roomID int64) *ChatRoom {
	crm.rMutex.RLock()
	room := crm.Rooms[roomID]
	crm.rMutex.RUnlock()
	if room != nil {
		return room
	}

	crm.rMutex.Lock()
	defer crm.rMutex.Unlock()

	room = crm.Rooms[roomID]
	if room != nil {
		return room
	}

	room = NewChatRoom(roomID, "")
	crm.Rooms[roomID] = room

	return room
}

func (crm *ChatRoomManager) SubscribeRoom(player *player.Player, roomID int64) {
	room := crm.GetRoomNotNil(roomID)
	room.Join(player)

	player.SendGameData(proto.SubscribeRoomAck, &proto.SubscribeRoomResp{
		RoomID: roomID,
	})
}

func (crm *ChatRoomManager) UnsubscribeRoom(player *player.Player, roomID int64) {
	room := crm.GetRoomNotNil(roomID)
	room.Leave(player)

	player.SendGameData(proto.UnsubscribeRoomAck, &proto.UnsubscribeRoomResp{
		RoomID: roomID,
	})
}

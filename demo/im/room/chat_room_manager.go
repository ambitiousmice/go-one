package room

import "sync"

var CRM *ChatRoomManager

type ChatRoomManager struct {
	rMutex sync.RWMutex
	Rooms  map[int64]*ChatRoom
}

func (crm *ChatRoomManager) GetRoom(roomID int64) *ChatRoom {
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

	NewChatRoom(roomID, "")
	crm.Rooms[roomID] = room

	return room
}

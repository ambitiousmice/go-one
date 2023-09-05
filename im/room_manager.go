package im

import "sync"

type RoomManager struct {
	rwMutex sync.RWMutex
	rooms   map[int32]*Room
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[int32]*Room),
	}
}

func (rm *RoomManager) GetRoom(roomID int32) *Room {
	rm.rwMutex.RLock()
	defer rm.rwMutex.RUnlock()
	return rm.rooms[roomID]
}

func (rm *RoomManager) AddRoom(room *Room) {
	rm.rwMutex.Lock()
	defer rm.rwMutex.Unlock()
	rm.rooms[room.ID] = room
}

func (rm *RoomManager) RemoveRoom(roomID int32) {
	rm.rwMutex.Lock()
	defer rm.rwMutex.Unlock()
	delete(rm.rooms, roomID)
}

func (rm *RoomManager) GetRooms() map[int32]*Room {
	rm.rwMutex.RLock()
	defer rm.rwMutex.RUnlock()
	return rm.rooms
}

func (rm *RoomManager) GetRoomCount() int {
	return len(rm.rooms)
}

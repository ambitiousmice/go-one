package game

import (
	"math/rand"
	"reflect"
	"sync"
)

type RoomManager struct {
	mutex            sync.RWMutex
	roomType         string
	roomMaxPlayerNum int
	roomIDStart      int64
	roomIDEnd        int64
	matchStrategy    string
	IDPool           *RoomIDPool
	rooms            map[int64]*Room
	roomJoinOrder    []int64
}

func NewRoomManager(roomType string, roomMaxPlayerNum int, roomIDStart int64, roomIDEnd int64, matchStrategy string) *RoomManager {
	idPool, err := NewRoomIDPool(roomIDStart, roomIDEnd)
	if err != nil {
		panic("init room id pool error: " + err.Error())
	}

	return &RoomManager{
		roomType:         roomType,
		roomIDStart:      roomIDStart,
		roomIDEnd:        roomIDEnd,
		matchStrategy:    matchStrategy,
		roomMaxPlayerNum: roomMaxPlayerNum,
		IDPool:           idPool,
		rooms:            make(map[int64]*Room),
		roomJoinOrder:    make([]int64, 0),
	}
}

func (rm *RoomManager) GetRoom(roomID int64) *Room {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.rooms[roomID]
}

func (rm *RoomManager) GetRoomByStrategy() *Room {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	switch rm.matchStrategy {
	case RoomStrategyOrder:
		return rm.matchRoomByOrder()
	case RoomStrategyRandom:
		return rm.matchRoomRandomly()
	case RoomStrategyBalanced:
		return rm.matchRoomBalanced()
	default:
		return rm.matchRoomByOrder()
	}
}

func (rm *RoomManager) matchRoomByOrder() *Room {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	// 遍历房间加入的顺序
	for _, roomID := range rm.roomJoinOrder {
		room := rm.rooms[roomID]
		if room != nil && room.GetPlayerCount() < room.MaxPlayerNum {
			return room
		}
	}

	// 如果没有可用房间，创建一个新房间
	return rm.createRoom()
}

func (rm *RoomManager) matchRoomRandomly() *Room {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	// 随机打乱房间加入顺序
	rand.Shuffle(len(rm.roomJoinOrder), func(i, j int) {
		rm.roomJoinOrder[i], rm.roomJoinOrder[j] = rm.roomJoinOrder[j], rm.roomJoinOrder[i]
	})

	return rm.matchRoomByOrder()
}

func (rm *RoomManager) matchRoomBalanced() *Room {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	// 寻找最少人数的房间
	var minRoom *Room
	minPlayers := 99999 // 初始值设置为一个较大的数

	for _, roomID := range rm.roomJoinOrder {
		room := rm.rooms[roomID]
		if room != nil {
			playerCount := room.GetPlayerCount()
			if playerCount < room.MaxPlayerNum && playerCount < minPlayers {
				minPlayers = playerCount
				minRoom = room
			}
		}
	}

	// 如果没有可用房间，创建一个新房间
	if minRoom == nil {
		return rm.createRoom()
	}

	return minRoom
}

func (rm *RoomManager) createRoom() *Room {
	roomID, err := rm.IDPool.Get()
	if err != nil {
		return nil // ID 池已用尽
	}

	roomObjType := gameServer.getRoomObjType(rm.roomType)
	iRoomValue := reflect.New(roomObjType)
	iRoom := iRoomValue.Interface().(IRoom)

	room := reflect.Indirect(iRoomValue).FieldByName("Room").Addr().Interface().(*Room)
	room.I = iRoom
	room.init(roomID, rm.roomType, rm.roomMaxPlayerNum)

	rm.rooms[room.ID] = room
	rm.roomJoinOrder = append(rm.roomJoinOrder, roomID)

	return room
}

func (rm *RoomManager) RemoveRoom(roomID int64) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	delete(rm.rooms, roomID)
}

func (rm *RoomManager) GetRooms() map[int64]*Room {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.rooms
}

func (rm *RoomManager) GetRoomCount() int {
	return len(rm.rooms)
}

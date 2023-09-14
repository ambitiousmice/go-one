package game

import "go-one/common/log"

type RoomLobby struct {
	Room
}

func (r *RoomLobby) GetRoomType() string {
	return ROOM_LOBBY
}

func (r *RoomLobby) OnCreated() {
	log.Info("RoomLobby created")
}

func (r *RoomLobby) OnDestroyed() {
	log.Info("RoomLobby destroyed")
}

func (r *RoomLobby) OnJoined(player *Player) {
	log.Info("RoomLobby joined")
}

func (r *RoomLobby) OnLeft(player *Player) {
	log.Info("RoomLobby left")
}

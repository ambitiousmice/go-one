package game

import (
	"go-one/common/log"
	"go-one/common/proto"
)

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
	joinRoomResp := &proto.JoinRoomResp{
		RoomID:   r.ID,
		RoomType: r.Type,
	}

	player.SendGameData(proto.JoinRoomFromGame, joinRoomResp)
}

func (r *RoomLobby) OnLeft(player *Player) {
	log.Info("RoomLobby left")
}

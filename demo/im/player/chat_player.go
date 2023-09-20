package player

import (
	"go-one/demo/im/proto"
	"go-one/demo/im/room"
	"go-one/game"
)

type ChatPlayer struct {
	*game.Player

	subscribeRooms map[int64]*room.ChatRoom
}

func (p *ChatPlayer) OnCreated() {

}

func (p *ChatPlayer) OnDestroy() {

}

func (p *ChatPlayer) OnClientConnected() {

}

func (p *ChatPlayer) OnClientDisconnected() {

}

func (p *ChatPlayer) SubscribeRoom(room *room.ChatRoom) {
	p.subscribeRooms[room.ID] = room

	p.SendGameData(proto.SubscribeRoomAck, &proto.SubscribeRoomResp{
		RoomID: room.ID,
	})
}

func (p *ChatPlayer) UnSubscribeRoom(room *room.ChatRoom) {
	delete(p.subscribeRooms, room.ID)
	p.SendGameData(proto.UnSubscribeRoomAck, &proto.UnSubscribeRoomResp{
		RoomID: room.ID,
	})
}

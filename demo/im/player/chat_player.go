package player

import (
	"go-one/demo/im/proto"
	"go-one/demo/im/room"
	"go-one/game/player"
	"go-one/game/scene_center"
)

type ChatPlayer struct {
	player.Player

	subscribeRooms map[int64]*room.ChatRoom
}

func (p *ChatPlayer) OnCreated() {
	p.subscribeRooms = make(map[int64]*room.ChatRoom)
}

func (p *ChatPlayer) OnDestroy() {

}

func (p *ChatPlayer) OnClientConnected() {

}

func (p *ChatPlayer) OnClientDisconnected() {
	for _, r := range p.subscribeRooms {
		r.Leave(&p.Player)
	}
	scene_center.Leave(&p.Player)
}

func (p *ChatPlayer) SubscribeRoom(room *room.ChatRoom) {
	p.subscribeRooms[room.ID] = room

	p.SendGameData(proto.SubscribeRoomAck, &proto.SubscribeRoomResp{
		RoomID: room.ID,
	})
}

func (p *ChatPlayer) UnSubscribeRoom(room *room.ChatRoom) {
	delete(p.subscribeRooms, room.ID)
	p.SendGameData(proto.UnsubscribeRoomAck, &proto.UnsubscribeRoomResp{
		RoomID: room.ID,
	})
}

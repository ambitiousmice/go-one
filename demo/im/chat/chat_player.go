package chat

import (
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/demo/im/proto"
	"github.com/ambitiousmice/go-one/game/entity"
)

type ChatPlayer struct {
	entity.Player

	subscribeRooms map[int64]*ChatRoom
}

func (p *ChatPlayer) OnCreated() {
	p.subscribeRooms = make(map[int64]*ChatRoom)
}

func (p *ChatPlayer) OnDestroy() {

}

func (p *ChatPlayer) OnClientDisconnected() {
	for _, r := range p.subscribeRooms {
		r.Leave(&p.Player)
	}
	p.LeaveScene()
}

func (p *ChatPlayer) OnJoinScene() {

	log.Infof("%s join %s", p, p.Scene)
}

func (p *ChatPlayer) SubscribeRoom(room *ChatRoom) {
	p.subscribeRooms[room.ID] = room

	p.SendGameData(proto.SubscribeRoomAck, &proto.SubscribeRoomResp{
		RoomID: room.ID,
	})
}

func (p *ChatPlayer) UnSubscribeRoom(room *ChatRoom) {
	delete(p.subscribeRooms, room.ID)
	p.SendGameData(proto.UnsubscribeRoomAck, &proto.UnsubscribeRoomResp{
		RoomID: room.ID,
	})
}

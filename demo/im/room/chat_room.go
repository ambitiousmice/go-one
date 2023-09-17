package room

import (
	"go-one/common/log"
	"go-one/common/proto"
	"go-one/demo/im/common"
	"go-one/game"
)

type ChatRoom struct {
	game.Scene
}

func (r *ChatRoom) GetSceneType() string {
	return common.RoomTypeChat
}

func (r *ChatRoom) OnCreated() {
	log.Infof("room created,%s", r.String())
}

func (r *ChatRoom) OnDestroyed() {
	log.Infof("room destroyed,%s", r.String())
}

func (r *ChatRoom) OnJoined(player *game.Player) {
	log.Info("player joined room,%s | %s", player.String(), r.String())
	joinRoomResp := &proto.JoinSceneResp{
		SceneID:   r.ID,
		SceneType: r.Type,
	}

	player.SendGameData(proto.JoinScene, joinRoomResp)
}

func (r *ChatRoom) OnLeft(player *game.Player) {
	r.RemovePlayer(player)
}

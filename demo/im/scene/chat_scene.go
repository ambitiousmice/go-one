package scene

import (
	"go-one/common/log"
	"go-one/common/proto"
	"go-one/demo/im/common"
	"go-one/demo/im/room"
	"go-one/game"
)

type ChatScene struct {
	game.Scene

	RoomManager *room.ChatRoomManager
}

func (r *ChatScene) GetSceneType() string {
	return common.SceneTypeChat
}

func (r *ChatScene) OnCreated() {
	log.Infof("scene created,%s", r.String())
}

func (r *ChatScene) OnDestroyed() {
	log.Infof("scene destroyed,%s", r.String())
}

func (r *ChatScene) OnJoined(player *game.Player) {
	log.Info("player joined scene,%s | %s", player.String(), r.String())
	joinSceneResp := &proto.JoinSceneResp{
		SceneID:   r.ID,
		SceneType: r.Type,
	}

	player.SendGameData(proto.JoinScene, joinSceneResp)
}

func (r *ChatScene) OnLeft(player *game.Player) {
	r.RemovePlayer(player)
}

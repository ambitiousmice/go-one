package scene_center

import (
	"go-one/common/common_proto"
	"go-one/common/log"
	"go-one/game/common"
	"go-one/game/player"
)

type LobbyScene struct {
	Scene
}

func (r *LobbyScene) GetSceneType() string {
	return common.SceneTypeLobby
}

func (r *LobbyScene) OnCreated() {
	log.Info("LobbyScene created")
}

func (r *LobbyScene) OnDestroyed() {
	log.Info("LobbyScene destroyed")
}

func (r *LobbyScene) OnJoined(player *player.Player) {
	joinSceneResp := &common_proto.JoinSceneResp{
		SceneID:   r.ID,
		SceneType: r.Type,
	}

	log.Infof("%s joined lobby ", player.String())

	player.SendGameData(common_proto.JoinSceneAck, joinSceneResp)
}

func (r *LobbyScene) OnLeft(player *player.Player) {
	log.Infof("%s left lobby ", player.String())
}

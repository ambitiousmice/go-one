package game

import (
	"go-one/common/log"
	"go-one/common/proto"
)

type SceneLobby struct {
	Scene
}

func (r *SceneLobby) GetSceneType() string {
	return SceneTypeLobby
}

func (r *SceneLobby) OnCreated() {
	log.Info("SceneLobby created")
}

func (r *SceneLobby) OnDestroyed() {
	log.Info("SceneLobby destroyed")
}

func (r *SceneLobby) OnJoined(player *Player) {
	joinSceneResp := &proto.JoinSceneResp{
		SceneID:   r.ID,
		SceneType: r.Type,
	}

	log.Infof("SceneLobby joined, player=<%s>", player.String())

	player.SendGameData(proto.JoinSceneAck, joinSceneResp)
}

func (r *SceneLobby) OnLeft(player *Player) {
	log.Info("SceneLobby left")
}

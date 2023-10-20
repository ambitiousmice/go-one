package entity

import (
	"go-one/common/common_proto"
	"go-one/common/log"
	"go-one/demo/multiplayer/common"
	"go-one/game/entity"
)

type MultiplayerScene struct {
	entity.Scene
}

func (r *MultiplayerScene) GetSceneType() string {
	return common.SceneTypeMultiplayer
}

func (r *MultiplayerScene) OnCreated() {

}

func (r *MultiplayerScene) OnDestroyed() {
	log.Infof("%s destroyed", r)
}

func (r *MultiplayerScene) OnJoined(p *entity.Player) {
	log.Infof("%s joined %s ", p, r)
	joinSceneResp := &common_proto.JoinSceneResp{
		SceneID:   r.ID,
		SceneType: r.Type,
	}

	p.SendGameData(common_proto.JoinScene, joinSceneResp)
}

func (r *MultiplayerScene) OnLeft(p *entity.Player) {

}

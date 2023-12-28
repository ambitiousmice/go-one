package entity

import (
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/game/common"
)

type LobbyScene struct {
	Scene
}

func (l *LobbyScene) GetSceneType() string {
	return common.SceneTypeLobby
}

func (l *LobbyScene) OnCreated() {
	log.Infof("lobby scene:%d created", l.Scene.ID)
}

func (l *LobbyScene) OnDestroyed() {
	log.Infof("lobby scene:%d  destroyed", l.Scene.ID)
}

func (l *LobbyScene) OnJoined(player *Player) {
	joinSceneResp := &common_proto.JoinSceneResp{
		SceneID:   l.ID,
		SceneType: l.Type,
	}

	log.Infof("%s joined lobby:%d ", player.String(), l.Scene.ID)

	player.SendGameData(common_proto.JoinSceneAck, joinSceneResp)
}

func (l *LobbyScene) OnLeft(player *Player) {
	log.Infof("%s left lobby:%d", player.String(), l.Scene.ID)
}

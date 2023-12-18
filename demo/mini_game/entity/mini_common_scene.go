package entity

import (
	"go-one/common/common_proto"
	"go-one/common/log"
	"go-one/demo/mini_game/common"
	"go-one/game/entity"
)

type MiniCommonScene struct {
	entity.Scene
}

func (m *MiniCommonScene) GetSceneType() string {
	return common.SceneTypeMiniCommon
}

func (m *MiniCommonScene) OnCreated() {

}

func (m *MiniCommonScene) OnDestroyed() {
	log.Infof("%s destroyed", m)
}

func (m *MiniCommonScene) OnJoined(p *entity.Player) {
	log.Infof("%s joined %s ", p, m)
	joinSceneResp := &common_proto.JoinSceneResp{
		SceneID:   m.ID,
		SceneType: m.Type,
	}

	p.SendGameData(common_proto.JoinScene, joinSceneResp)
}

func (m *MiniCommonScene) OnLeft(p *entity.Player) {

}

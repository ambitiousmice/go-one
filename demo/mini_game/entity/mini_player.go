package entity

import (
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/game/entity"
)

type MiniPlayer struct {
	entity.Player
}

func (p *MiniPlayer) OnCreated() {

}

func (p *MiniPlayer) OnDestroy() {

}

func (p *MiniPlayer) OnClientConnected() {

}

func (p *MiniPlayer) OnClientDisconnected() {

	p.LeaveScene()
}

func (p *MiniPlayer) OnJoinScene() {

	log.Infof("%s join %s", p, p.Scene)
}

package entity

import (
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/game/entity"
)

type Player struct {
	entity.Player
}

func (p *Player) OnCreated() {

}

func (p *Player) OnDestroy() {

}

func (p *Player) OnClientConnected() {

}

func (p *Player) OnClientDisconnected() {

	p.LeaveScene()
}

func (p *Player) UpdateData() {

}

func (p *Player) OnJoinScene() {

	log.Infof("%s join %s", p, p.Scene)
}

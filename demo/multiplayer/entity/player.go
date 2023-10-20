package entity

import (
	"go-one/common/log"
	"go-one/game/entity"
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

func (p *Player) OnJoinScene() {

	log.Infof("%s join %s", p, p.Scene)
}

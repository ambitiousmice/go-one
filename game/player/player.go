package player

type Player struct {
	*BasePlayer

	I IPlayer
}

func (p *Player) init(entityID int64, gateID uint8) {
	p.BasePlayer = NewBasePlayer(entityID, gateID)

	p.I.OnCreated()
}

package player

type Player struct {
	*BasePlayer

	I IPlayer
}

func (p *Player) init(entityID int64, gateClusterID uint8) {
	p.BasePlayer = NewBasePlayer(entityID, gateClusterID)

	p.I.OnCreated()
}

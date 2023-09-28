package processor_center

import (
	"go-one/game/player"
)

type Processor interface {
	Process(player *player.Player, param []byte)
	GetCmd() uint16
}

package processor_center

import (
	"go-one/game/entity"
)

type Processor interface {
	Process(player *entity.Player, param []byte)
	GetCmd() uint16
}

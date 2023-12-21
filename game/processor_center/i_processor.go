package processor_center

import (
	"github.com/ambitiousmice/go-one/game/entity"
)

type Processor interface {
	Process(player *entity.Player, param []byte)
	GetCmd() uint16
}

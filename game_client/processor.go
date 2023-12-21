package game_client

import (
	"github.com/ambitiousmice/go-one/common/log"
	"strconv"
)

var ProcessorContext = make(map[uint16]Processor)

type Processor interface {
	Process(client *Client, param []byte)
	GetCmd() uint16
}

func RegisterProcessor(p Processor) {
	processor := ProcessorContext[p.GetCmd()]
	if processor != nil {
		log.Panic("duplicate processor_center: " + strconv.Itoa(int(p.GetCmd())))
	}
	ProcessorContext[p.GetCmd()] = p
}

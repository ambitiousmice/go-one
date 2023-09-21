package game_client

import "strconv"

var ProcessorContext = make(map[uint16]Processor)

type Processor interface {
	Process(client *Client, param []byte)
	GetCmd() uint16
}

func RegisterProcessor(p Processor) {
	processor := ProcessorContext[p.GetCmd()]
	if processor != nil {
		panic("duplicate processor: " + strconv.Itoa(int(p.GetCmd())))
	}
	ProcessorContext[p.GetCmd()] = p
}

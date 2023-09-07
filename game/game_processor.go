package game

import (
	"go-one/common/log"
	"go-one/common/proto"
)

var processContext = make(map[uint16]Processor)

type Processor interface {
	Process(basePlayer *BasePlayer, param []byte)
	GetCmd() uint16
}

func RegisterProcessor(p Processor) {
	processContext[p.GetCmd()] = p
}

func gameProcess(gp *GateProxy, entityID int64, req *proto.GameReq) {
	basePlayer := GetPlayer(entityID)
	if basePlayer == nil {
		log.Warnf("player:<%d> not found", entityID)
		basePlayer = NewBasePlayer(entityID)
		basePlayer.gateProxy = gp
		AddPlayer(basePlayer)
	}
	processor := processContext[req.Cmd]
	if processor == nil {
		log.Errorf("player%d send invalid cmd: %d", entityID, req.Cmd)
		basePlayer.SendCommonErrorMsg("invalid cmd")
		return
	}
	processor.Process(basePlayer, req.Param)
}

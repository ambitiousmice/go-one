package game

import (
	"go-one/common/log"
	"go-one/common/proto"
)

var processContext = make(map[uint16]Processor)

type Processor interface {
	Process(basePlayer *Player, param []byte)
	GetCmd() uint16
}

func RegisterProcessor(p Processor) {
	processContext[p.GetCmd()] = p
}

func gameProcess(gp *GateProxy, entityID int64, req *proto.GameReq) {
	player := GetPlayer(entityID)
	if player == nil {
		log.Warnf("player:<%d> not found", entityID)
		player = AddPlayer(entityID, gp.gateID)
		player.UpdateStatus(PlayerStatusOnline)
	}
	processor := processContext[req.Cmd]
	if processor == nil {
		log.Errorf("player%d send invalid cmd: %d", entityID, req.Cmd)
		player.SendCommonErrorMsg("invalid cmd")
		return
	}
	processor.Process(player, req.Param)
}

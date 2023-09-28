package base_processor

import (
	"go-one/common/consts"
	"go-one/common/pktconn"
	"go-one/game/player"
)

func UnPackMsg(player *player.Player, param []byte, obj any) {
	err := pktconn.MSG_PACKER.UnpackMsg(param, obj)
	if err != nil {
		player.SendCommonErrorMsg(consts.ParamError)
	}
}

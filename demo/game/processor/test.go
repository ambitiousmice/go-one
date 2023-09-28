package processor

import (
	"go-one/common/common_proto"
	"go-one/common/log"
	"go-one/game/player"
)

type Test struct {
}

func (t *Test) Process(player *player.Player, param []byte) {
	log.Infof("test process: %s", string(param))
	player.SendGameMsg(&common_proto.GameResp{
		Cmd:  1,
		Data: []byte("test"),
	})
}

func (t *Test) GetCmd() uint16 {
	return 1
}

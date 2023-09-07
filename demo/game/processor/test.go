package processor

import (
	"go-one/common/log"
	"go-one/common/proto"
	"go-one/game"
)

type Test struct {
}

func (t *Test) Process(basePlayer *game.BasePlayer, param []byte) {
	log.Infof("test process: %s", string(param))
	basePlayer.SendGameMsg(&proto.GameResp{
		Cmd:  1,
		Data: []byte("test"),
	})
}

func (t *Test) GetCmd() uint16 {
	return 1
}

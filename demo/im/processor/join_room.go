package processor

import (
	"go-one/common/proto"
	"go-one/game"
)

type JoinRoom struct {
}

func (t *JoinRoom) Process(basePlayer *game.BasePlayer, param []byte) {

	basePlayer.SendGameMsg(&proto.GameResp{
		Cmd:  1,
		Data: []byte("test"),
	})
}

func (t *JoinRoom) GetCmd() uint16 {
	return 1
}

package game_client

import (
	"go-one/common/proto"
)

type JoinSceneProcessor struct {
}

func (p *JoinSceneProcessor) Process(client *Client, param []byte) {
	joinSceneResp := &proto.JoinSceneResp{}
	UnPackMsg(param, joinSceneResp)
	client.I.OnJoinScene(client, joinSceneResp)
}

func (p *JoinSceneProcessor) GetCmd() uint16 {
	return proto.JoinSceneAck
}

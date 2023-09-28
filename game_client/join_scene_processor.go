package game_client

import (
	"go-one/common/common_proto"
)

type JoinSceneProcessor struct {
}

func (p *JoinSceneProcessor) Process(client *Client, param []byte) {
	joinSceneResp := &common_proto.JoinSceneResp{}
	UnPackMsg(param, joinSceneResp)
	client.I.OnJoinScene(client, joinSceneResp)
}

func (p *JoinSceneProcessor) GetCmd() uint16 {
	return common_proto.JoinSceneAck
}

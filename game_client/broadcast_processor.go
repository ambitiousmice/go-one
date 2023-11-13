package game_client

import (
	"go-one/common/common_proto"
	"go-one/common/log"
	"go-one/game/common"
)

type BroadcastProcessor struct {
}

func (p *BroadcastProcessor) Process(client *Client, param []byte) {
	broadcastMsg := &common_proto.GateBroadcastMsg{}
	err := common.UnPackMsg(param, broadcastMsg)
	if err != nil {
		log.Errorf("unpack msg error: %s", err.Error())
		return
	}
	log.Infof("收到广播消息:%s", broadcastMsg.Data)
}

func (p *BroadcastProcessor) GetCmd() uint16 {
	return common_proto.BroadcastFromServer
}

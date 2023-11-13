package gate

import (
	"github.com/IBM/sarama"
	"go-one/common/common_proto"
	"go-one/common/log"
	"go-one/common/pktconn"
)

func BroadcastMsgHandler(msg *sarama.ConsumerMessage) {
	m := &common_proto.GateBroadcastMsg{}
	err := pktconn.MSG_PACKER.UnpackMsg(msg.Value, m)
	if err != nil {
		log.Errorf("unpack gate kafka message error(%v)", err)
		return
	}

	gateServer.Broadcast(m)
}

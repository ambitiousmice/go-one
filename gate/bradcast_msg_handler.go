package gate

import (
	"github.com/IBM/sarama"
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/pktconn"
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

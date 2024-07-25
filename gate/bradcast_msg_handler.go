package gate

import (
	"github.com/IBM/sarama"
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/pktconn"
	"github.com/ambitiousmice/go-one/common/utils"
	"github.com/ambitiousmice/go-one/gate/mq"
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

func ReceiveLoginSyncNotifyHandler(msg *sarama.ConsumerMessage) {
	m := &mq.GateLoginSyncNotify{}
	jsonStruct, err := utils.ByteArrayToJsonStruct(msg.Value, m)
	if err != nil {
		log.Errorf("unpack gate kafka message error(%v)", err)
		return
	}
	log.Infof("mq topic:GateSyncPlayer resolve result：%s", jsonStruct)

	gateServer := GetGateServer()
	if gateServer != nil {
		oldCP := gateServer.getClientProxy(m.EntityID)
		if oldCP != nil && oldCP.clientID != m.ClientID {
			//
			log.Infof("异地登陆链接断开：%s", jsonStruct)
			oldCP.CloseAll()
		}
	}

}

package kafka

import (
	"github.com/ambitiousmice/go-one/common/mq/kafka"
	"github.com/ambitiousmice/go-one/gate/mq"
)

func SendGateLoginSyncNotify(entityID int64, clientID string) {
	if kafka.Producer == nil {
		return
	}
	message := &mq.GateLoginSyncNotify{
		EntityID: entityID,
		ClientID: clientID,
	}
	kafka.Producer.SendMessage(mq.GateSyncPlayer, "GateLoginSyncNotify", message) // 创建 Kafka 消息
}

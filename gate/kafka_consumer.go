package gate

import (
	"github.com/IBM/sarama"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/mq/kafka"
	"github.com/ambitiousmice/go-one/gate/mq"
)

func init() {
	kafka.RegisterConsumerHandler(consts.GateBroadcastTopic, &Consumer{})
}

type Consumer struct {
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Infof("message channel was closed")
				return nil
			}

			switch message.Topic {
			case consts.GateBroadcastTopic:
				BroadcastMsgHandler(message)
			case mq.GateSyncPlayer:
				//找到用户
				ReceiveLoginSyncNotifyHandler(message)
			}

			log.Infof("gate kafka message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

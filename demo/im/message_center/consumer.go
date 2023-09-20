package message_center

import (
	"github.com/IBM/sarama"
	"go-one/common/log"
	"go-one/common/pktconn"
)

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

			chatMessage := &ChatMessage{}
			err := pktconn.MSG_PACKER.UnpackMsg(message.Value, chatMessage)
			if err != nil {
				log.Errorf("unpack chat message error(%v)", err)
				continue
			}

			switch message.Topic {
			case One:
				OneMessageHandler(chatMessage)
			case Room:
				RoomMessageHandler(chatMessage)
			case Broadcast:
				BroadcastMessageHandler(chatMessage)
			}
			log.Infof("ChatMessage claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

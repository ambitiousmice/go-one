package kafka

import (
	"github.com/IBM/sarama"
	"go-one/common/log"
	"go-one/common/pktconn"
	"strings"
)

type IProducer interface {
	SendMessage(topic, key string, value any)
	SendMessages(msg []*sarama.ProducerMessage)
	Close()
}

type SyncProducer struct {
	producer sarama.SyncProducer
}

func NewSyncProducer(producerConfig ProducerConfig) IProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = producerConfig.RetryMax
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	syncProducer, err := sarama.NewSyncProducer(strings.Split(producerConfig.Brokers, ","), config)
	if err != nil {
		log.Panic(err)
	}

	return &SyncProducer{
		producer: syncProducer,
	}
}

func (p *SyncProducer) SendMessage(topic, key string, value any) {
	byteValue, err := pktconn.MSG_PACKER.PackMsg(value, nil)
	if err != nil {
		log.Warnf("pack message error(%v)", err)
		return
	}

	msg := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(key),
		Topic: topic,
		Value: sarama.ByteEncoder(byteValue),
	}
	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		log.Warnf("send message error(%v)", err)
		return
	}
}

func (p *SyncProducer) SendMessages(msg []*sarama.ProducerMessage) {
	err := p.producer.SendMessages(msg)
	if err != nil {
		log.Warnf("pack message error(%v)", err)
		return
	}
}

func (p *SyncProducer) Close() {
	p.producer.Close()
}

type AsyncProducer struct {
	producer sarama.AsyncProducer
}

func NewAsyncProducer(producerConfig ProducerConfig) IProducer {
	config := sarama.NewConfig()
	config.Producer.Retry.Max = producerConfig.RetryMax
	config.Producer.Return.Successes = false
	config.Producer.Return.Errors = true
	syncProducer, err := sarama.NewAsyncProducer(strings.Split(producerConfig.Brokers, ","), config)
	if err != nil {
		log.Panic(err)
	}

	return &AsyncProducer{
		producer: syncProducer,
	}
}

func (p *AsyncProducer) SendMessage(topic, key string, value any) {
	byteValue, err := pktconn.MSG_PACKER.PackMsg(value, nil)
	if err != nil {
		log.Warnf("pack message error(%v)", err)
		return
	}
	msg := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(key),
		Topic: topic,
		Value: sarama.ByteEncoder(byteValue),
	}
	p.producer.Input() <- msg
}

func (p *AsyncProducer) SendMessages(messages []*sarama.ProducerMessage) {
	for _, msg := range messages {
		p.producer.Input() <- msg
	}
}

func (p *AsyncProducer) Close() {
	p.producer.Close()
}

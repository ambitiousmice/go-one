package kafka

import "github.com/IBM/sarama"

type IProducer interface {
	SendMessage(msg *sarama.ProducerMessage)
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
	syncProducer, err := sarama.NewSyncProducer(producerConfig.Brokers, config)
	if err != nil {
		panic(err)
	}

	return &SyncProducer{
		producer: syncProducer,
	}
}

func (p *SyncProducer) SendMessage(msg *sarama.ProducerMessage) {
	p.producer.SendMessage(msg)
}

func (p *SyncProducer) SendMessages(msg []*sarama.ProducerMessage) {
	p.producer.SendMessages(msg)
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
	syncProducer, err := sarama.NewAsyncProducer(producerConfig.Brokers, config)
	if err != nil {
		panic(err)
	}

	return &AsyncProducer{
		producer: syncProducer,
	}
}

func (p *AsyncProducer) SendMessage(msg *sarama.ProducerMessage) {
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

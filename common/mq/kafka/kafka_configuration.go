package kafka

import "github.com/IBM/sarama"

type ProducerConfig struct {
	Brokers  []string `yaml:"brokers"`
	RetryMax int      `yaml:"retry-max"`
	Version  string   `yaml:"version"`
	Sync     bool     `yaml:"sync"`
}

type ConsumerConfig struct {
	Brokers []string `yaml:"brokers"`
	Group   string   `yaml:"group"`
	Version string   `yaml:"version"`
}

var producer IProducer
var consumer sarama.ConsumerGroup

func InitProducer(producerConfig ProducerConfig) {
	if len(producerConfig.Brokers) == 0 {
		return
	}

	if producerConfig.Sync {
		producer = NewSyncProducer(producerConfig)
	} else {
		producer = NewAsyncProducer(producerConfig)
	}
}

func InitConsumer(consumerConfig ConsumerConfig) {
	if len(consumerConfig.Brokers) == 0 {
		return
	}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(consumerConfig.Brokers, consumerConfig.Group, config)
	if err != nil {
		panic(err)
	}

	consumer = consumerGroup
}

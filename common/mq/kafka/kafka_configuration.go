package kafka

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"github.com/ambitiousmice/go-one/common/log"
	"strings"
)

var consumerHandlerContext = make(map[string]sarama.ConsumerGroupHandler)

type ProducerConfig struct {
	Brokers  string `yaml:"brokers"`
	RetryMax int    `yaml:"retry-max"`
	Version  string `yaml:"version"`
	Sync     bool   `yaml:"sync"`
}

type ConsumerConfig struct {
	HandlerName string `yaml:"handler-name"`
	Brokers     string `yaml:"brokers"`
	Group       string `yaml:"group"`
	Topics      string `yaml:"topics"`
	Version     string `yaml:"version"`
}

func RegisterConsumerHandler(name string, handler sarama.ConsumerGroupHandler) {
	consumerHandlerContext[name] = handler
}

var Producer IProducer

func InitProducer(producerConfig ProducerConfig) {
	if producerConfig.Brokers == "" {
		return
	}

	if producerConfig.Sync {
		Producer = NewSyncProducer(producerConfig)
	} else {
		Producer = NewAsyncProducer(producerConfig)
	}
}

func InitConsumer(consumerConfigs []ConsumerConfig) {
	if len(consumerConfigs) == 0 {
		return
	}

	for _, c := range consumerConfigs {
		config := sarama.NewConfig()
		config.Consumer.Return.Errors = true

		consumerGroup, err := sarama.NewConsumerGroup(strings.Split(c.Brokers, ","), c.Group, config)
		if err != nil {
			log.Panic(err)
		}

		topics := strings.Split(c.Topics, ",")

		consumerHandler := consumerHandlerContext[c.HandlerName]
		if consumerHandler == nil {
			log.Panic("consumer handler: " + c.HandlerName + " not found")
		}

		go func() {
			for {
				// `Consume` should be called inside an infinite loop, when a
				// server-side rebalance happens, the consumer session will need to be
				// recreated to get the new claims
				if err := consumerGroup.Consume(context.Background(), topics, consumerHandler); err != nil {
					if errors.Is(err, sarama.ErrClosedConsumerGroup) {
						log.Errorf("consumer group has been closed, will close the consumer group")
						return
					}
					log.Panicf("Error from consumer: %v", err)
				}
			}
		}()

		log.Infof("consumer group: %s, topics: %s, handler: %s started", c.Group, c.Topics, c.HandlerName)
	}

}

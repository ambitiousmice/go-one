package context

import (
	"go-one/common/mq/kafka"
	"go-one/common/register"
)

func Init() error {
	err := ReadYaml()
	if err != nil {
		return nil
	}

	register.Run(oneConfig.Nacos)

	InitIDGenerator(oneConfig.IDGeneratorConfig)

	kafka.InitProducer(oneConfig.KafkaProducerConfig)
	kafka.InitConsumer(oneConfig.KafkaConsumerConfig)

	return nil
}

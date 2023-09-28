package context

import (
	"go-one/common/log"
	"go-one/common/mq/kafka"
	"go-one/common/register"
	"math/rand"
	"time"
)

func Init() {
	err := ReadYaml()
	if err != nil {
		panic("read yaml error:" + err.Error())
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))

	log.InitLogger(&oneConfig.Logger)

	register.Run(oneConfig.Nacos)

	err = InitIDGenerator(oneConfig.IDGeneratorConfig)
	if err != nil {
		panic("init id generator error:" + err.Error())
	}

	kafka.InitProducer(oneConfig.KafkaProducerConfig)
	kafka.InitConsumer(oneConfig.KafkaConsumerConfigs)

	log.Info("context init success")
}

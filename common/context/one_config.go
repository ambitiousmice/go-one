package context

import (
	"go-one/common/cache"
	"go-one/common/db"
	"go-one/common/log"
	"go-one/common/mq/kafka"
	"go-one/common/register"
)

type OneConfig struct {
	Nacos                register.NacosConf     `yaml:"nacos"`
	Logger               log.Config             `yaml:"logger"`
	IDGeneratorConfig    IDGeneratorConfig      `yaml:"id_generator"`
	KafkaProducerConfig  kafka.ProducerConfig   `yaml:"kafka-producer"`
	KafkaConsumerConfigs []kafka.ConsumerConfig `yaml:"kafka-consumers"`
	MongoDBConfig        db.MongoDBConfig       `yaml:"mongodb"`
	RedisConfig          cache.RedisConfig      `yaml:"redis"`
	PprofHost            string                 `yaml:"pprof-host"`
}

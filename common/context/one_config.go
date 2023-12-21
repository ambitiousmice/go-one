package context

import (
	"github.com/ambitiousmice/go-one/common/cache"
	"github.com/ambitiousmice/go-one/common/db"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/mq/kafka"
	"github.com/ambitiousmice/go-one/common/pool"
	"github.com/ambitiousmice/go-one/common/register"
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
	PoolConfig           pool.Config            `yaml:"pool-config"`
}

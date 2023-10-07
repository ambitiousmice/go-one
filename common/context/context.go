package context

import (
	"github.com/robfig/cron/v3"
	"go-one/common/log"
	"go-one/common/mq/kafka"
	"go-one/common/register"
	"math/rand"
	"time"
)

var cronTab = cron.New(cron.WithSeconds())
var cronTaskMap = make(map[string]cron.EntryID)

func init() {
	cronTab.Start()
}

func Init() {
	err := InitConfig()
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

func AddCronTask(taskName string, spec string, method func()) error {
	taskID := cronTaskMap[taskName]
	if taskID != 0 {
		cronTab.Remove(taskID)
	}

	newTaskID, err := cronTab.AddFunc(spec, method)
	if err != nil {
		return err
	}

	cronTaskMap[taskName] = newTaskID

	return nil
}

func RemoveCronTask(taskName string) {
	taskID := cronTaskMap[taskName]
	if taskID != 0 {
		cronTab.Remove(taskID)
	}
}

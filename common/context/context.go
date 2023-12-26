package context

import (
	"github.com/ambitiousmice/go-one/common/cache"
	"github.com/ambitiousmice/go-one/common/db/mongo"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/mq/kafka"
	"github.com/ambitiousmice/go-one/common/pool"
	"github.com/ambitiousmice/go-one/common/pool/goroutine_pool"
	"github.com/ambitiousmice/go-one/common/register"
	"github.com/robfig/cron/v3"
	"math/rand"
	"net/http"
	"time"
)

var cronTab = cron.New(cron.WithSeconds())
var cronTaskMap = make(map[string]cron.EntryID)
var configFromNacos string

func init() {
	cronTab.Start()
}

func Init() {
	err := InitConfig()
	if err != nil {
		log.Panic("read yaml error:" + err.Error())
	}
	rand.New(rand.NewSource(time.Now().UnixNano()))

	log.InitLogger(&oneConfig.Logger)

	register.Run(oneConfig.Nacos)

	InitConfigFromNacos()

	err = InitIDGenerator(oneConfig.IDGeneratorConfig)
	if err != nil {
		log.Panic("init id generator error:" + err.Error())
	}

	kafka.InitProducer(oneConfig.KafkaProducerConfig)
	kafka.InitConsumer(oneConfig.KafkaConsumerConfigs)

	mongo.InitMongo(&oneConfig.MongoDBConfig)

	cache.InitRedis(&oneConfig.RedisConfig)

	pool.InitPool(oneConfig.PoolConfig)

	if len(oneConfig.PprofHost) != 0 {
		log.Infof("run pprofServer ...")
		go setupPprofServer(oneConfig.PprofHost)
	}

	addTimerTask()

	log.Info("context init success")
}

func GetConfigFromNacos() string {
	return configFromNacos
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

func setupPprofServer(listenAddr string) {
	log.Infof("http server listening on %s", listenAddr)
	log.Infof("pprof http://%s/debug/pprof/ ... available commands: ", listenAddr)
	log.Infof("    go tool pprof http://%s/debug/pprof/heap", listenAddr)
	log.Infof("    go tool pprof http://%s/debug/pprof/profile", listenAddr)

	go func() {
		err := http.ListenAndServe(listenAddr, nil)
		if err != nil {
			log.Errorf("run pprofServer error:%s", err.Error())
		}
	}()
}

func addTimerTask() {
	if goroutine_pool.IsEnable() {
		err := AddCronTask("goroutine_pool_timer", "@every 30s", func() {
			log.Infof("goroutine pool running task num:%d", goroutine_pool.Running())
		})
		if err != nil {
			log.Panic("add goroutine_pool_timer error:" + err.Error())
		}
	}
}

package context

import (
	"github.com/robfig/cron/v3"
	"go-one/common/cache"
	"go-one/common/db"
	"go-one/common/log"
	"go-one/common/mq/kafka"
	"go-one/common/pool"
	"go-one/common/pool/goroutine_pool"
	"go-one/common/register"
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

	db.InitMongo(&oneConfig.MongoDBConfig)

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

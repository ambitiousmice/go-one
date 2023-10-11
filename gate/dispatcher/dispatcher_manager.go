package dispatcher

import (
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/robfig/cron/v3"
	"go-one/common/entity"
	"go-one/common/log"
	"go-one/common/pktconn"
	"go-one/common/register"
	"go-one/common/utils"
	"strconv"
	"sync/atomic"
)

var gameDispatcherMap = make(map[string]map[uint8]*GameDispatcher)

var gameLoadBalancerMap = make(map[string]LoadBalancer)

var crontab = *cron.New(cron.WithSeconds())

var gameDispatcherConfigs []entity.GameDispatcherConfig

var dispatcherClientPacketQueues []chan *pktconn.Packet
var dispatcherClientPacketQueuesIndex = new(uint64)

func InitGameDispatchers(dispatcherConfigs []entity.GameDispatcherConfig, queues []chan *pktconn.Packet) {
	if len(dispatcherConfigs) == 0 {
		log.Error("no game dispatcher config")
		return
	}
	gameDispatcherConfigs = dispatcherConfigs

	dispatcherClientPacketQueues = queues

	newGameDispatcher()

	crontab.AddFunc("@every 5s", func() {
		newGameDispatcher()
	})

	crontab.Start()
}

func newGameDispatcher() {
	for _, gameDispatcherConfig := range gameDispatcherConfigs {
		game := gameDispatcherConfig.Game
		groupName := gameDispatcherConfig.GroupName
		channelNum := gameDispatcherConfig.ChannelNum
		instances, err := register.NacosClient.SelectInstances(vo.SelectInstancesParam{
			ServiceName: game,
			GroupName:   groupName,
			HealthyOnly: true,
		})

		if err != nil {
			log.Warnf("select gameDispatcherConfig:< %s > server instances error: %s", gameDispatcherConfig, err.Error())
			continue
		}

		if len(instances) == 0 {
			log.Warnf("select gameDispatcherConfig:< %s > server instances is empty", gameDispatcherConfig)
			continue
		}

		checkMap := make(map[string]bool)
		for _, instance := range instances {
			clusterIDStr := instance.ClusterName

			if len(clusterIDStr) == 0 {
				panic("gameDispatcherConfig dispatcher instance clusterId is empty")
			}

			if checkMap[clusterIDStr] {
				panic("gameDispatcherConfig dispatcher instance gameClusterID is duplicate,ip:" + instance.Ip + ",port:" + utils.ToString(instance.Port))
			}
		}

		for _, instance := range instances {
			if gameDispatcherMap[game] == nil {
				gameDispatcherMap[game] = make(map[uint8]*GameDispatcher)
				gameLoadBalancerMap[game] = CreateLoadBalancer(gameDispatcherConfig.LoadBalancer)
			}

			clusterIDStr := instance.ClusterName

			clusterID, err := strconv.ParseUint(clusterIDStr, 10, 8)
			if err != nil {
				panic("gameDispatcherConfig dispatcher instance clusterId is not int")
			}

			gameDispatcher := gameDispatcherMap[game][uint8(clusterID)]
			if gameDispatcher != nil {
				continue
			}

			gameDispatcher = NewGameDispatcher(game, uint8(clusterID), instance.Ip, instance.Port)

			for i := uint8(0); i < channelNum; i++ {
				gameDispatcher.channels[i] = NewDispatcherChannel(i, gameDispatcher)
			}

			gameDispatcherMap[game][uint8(clusterID)] = gameDispatcher

			gameDispatcher.Run()
		}
	}
}

func ChooseGameDispatcher(game string, entityID int64) *GameDispatcher {
	loadBalancer := gameLoadBalancerMap[game]
	if loadBalancer == nil {
		log.Warnf("game:< %s > loadBalancer is nil", game)
		return nil
	}
	return loadBalancer.Choose(game, entityID)
}

func GetGameDispatcher(game string, gameClusterID uint8) *GameDispatcher {
	loadBalancer := gameLoadBalancerMap[game]
	if loadBalancer == nil {
		log.Warnf("game:< %s > loadBalancer is nil", game)
		return nil
	}
	return loadBalancer.FixedChoose(game, gameClusterID)
}

func getDispatcherClientPacketQueue() chan *pktconn.Packet {
	index := atomic.AddUint64(dispatcherClientPacketQueuesIndex, 1) % uint64(len(dispatcherClientPacketQueues))
	dispatcherClientPacketQueue := dispatcherClientPacketQueues[index]

	return dispatcherClientPacketQueue
}

package dispatcher

import (
	"github.com/ambitiousmice/go-one/common/entity"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/pktconn"
	"github.com/ambitiousmice/go-one/common/register"
	"github.com/ambitiousmice/go-one/common/utils"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/robfig/cron/v3"
	"strconv"
	"sync/atomic"
)

var gameDispatcherMap = make(map[string]map[uint8]*GameDispatcher)

var gameLoadBalancerMap = make(map[string]GameDispatcherLoadBalancer)

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
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("newGameDispatcher panic: %v", err)
		}
	}()
	for _, gameDispatcherConfig := range gameDispatcherConfigs {

		game := gameDispatcherConfig.Game
		groupName := gameDispatcherConfig.GroupName
		channelNum := gameDispatcherConfig.ChannelNum

		existDispatchers := gameDispatcherMap[game]

		instances, err := register.NacosClient.SelectInstances(vo.SelectInstancesParam{
			ServiceName: game,
			GroupName:   groupName,
			HealthyOnly: true,
		})

		if err != nil {
			log.Warnf("select gameDispatcherConfig:< %v > server instances error: %s", gameDispatcherConfig, err.Error())
			if err.Error() == "instance list is empty!" {
				if len(existDispatchers) != 0 {
					for clusterID, dispatcher := range existDispatchers {
						dispatcher.closeAll()
						delete(existDispatchers, clusterID)
					}
				}
			} else {
				continue
			}
		}

		checkMap := make(map[uint8]model.Instance)
		for _, instance := range instances {
			clusterIDStr := instance.ClusterName
			clusterID, err := strconv.ParseInt(clusterIDStr, 10, 8)
			if err != nil {
				log.Error("gameDispatcherConfig dispatcher instance clusterId is empty,ip:" + instance.Ip + ",port:" + utils.ToString(instance.Port))
				continue
			}
			clusterId := uint8(clusterID)
			_, exists := checkMap[clusterId]
			if exists {
				log.Error("gameDispatcherConfig dispatcher instance gameClusterID is duplicate,ip:" + instance.Ip + ",port:" + utils.ToString(instance.Port))
				continue
			}
			checkMap[clusterId] = instance
		}

		if len(existDispatchers) != 0 {
			for clusterID, dispatcher := range existDispatchers {
				exist := false
				for clusterId, _ := range checkMap {
					if clusterId == clusterID {
						exist = true
						delete(checkMap, clusterId)
						break
					}
				}
				if !exist {
					dispatcher.closeAll()
					delete(existDispatchers, clusterID)
				}
			}
		}

		for clusterID, instance := range checkMap {
			if gameDispatcherMap[game] == nil {
				gameDispatcherMap[game] = make(map[uint8]*GameDispatcher)
				gameLoadBalancerMap[game] = CreateLoadBalancer(gameDispatcherConfig.LoadBalancer)
			}

			gameDispatcher := gameDispatcherMap[game][clusterID]
			if gameDispatcher != nil {
				continue
			}

			gameDispatcher = NewGameDispatcher(game, clusterID, instance.Ip, instance.Port)

			for i := uint8(0); i < channelNum; i++ {
				gameDispatcher.channels[i] = NewDispatcherChannel(i, gameDispatcher)
			}

			gameDispatcher.Run()

			gameDispatcherMap[game][clusterID] = gameDispatcher

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

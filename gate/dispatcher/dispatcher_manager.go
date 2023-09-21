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
)

var gameDispatcherMap = make(map[string]map[uint8]*GameDispatcher)

var gameLoadBalancerMap = make(map[string]LoadBalancer)

var crontab = *cron.New(cron.WithSeconds())

var gameDispatcherConfigs []entity.GameDispatcherConfig

var dispatcherClientPacketQueue chan *pktconn.Packet

func InitGameDispatchers(dispatcherConfigs []entity.GameDispatcherConfig, queue chan *pktconn.Packet) {
	if len(dispatcherConfigs) == 0 {
		panic("no game dispatcher config")
	}
	gameDispatcherConfigs = dispatcherConfigs

	dispatcherClientPacketQueue = queue

	newGameDispatcher()

	crontab.AddFunc("@every 5s", func() {
		newGameDispatcher()
	})

	crontab.Start()
}

func newGameDispatcher() {
	for _, gameDispatcherConfig := range gameDispatcherConfigs {
		game := gameDispatcherConfig.Game
		channelNum := gameDispatcherConfig.ChannelNum
		instances, err := register.NacosClient.SelectInstances(vo.SelectInstancesParam{
			ServiceName: game,
			HealthyOnly: true,
		})

		if err != nil {
			log.Warnf("select gameDispatcherConfig:< %s > server instances error: ", gameDispatcherConfig, err.Error())
			continue
		}

		if len(instances) == 0 {
			log.Warnf("select gameDispatcherConfig:< %s > server instances is empty", gameDispatcherConfig)
			continue
		}

		checkMap := make(map[string]bool)
		for _, instance := range instances {
			gameIDStr := instance.Metadata["clusterId"]

			if len(gameIDStr) == 0 {
				panic("gameDispatcherConfig dispatcher instance clusterId is empty")
			}

			if checkMap[gameIDStr] {
				panic("gameDispatcherConfig dispatcher instance clusterId is duplicate,ip:" + instance.Ip + ",port:" + utils.ToString(instance.Port))
			}
		}

		for _, instance := range instances {
			if gameDispatcherMap[game] == nil {
				gameDispatcherMap[game] = make(map[uint8]*GameDispatcher)
				gameLoadBalancerMap[game] = CreateLoadBalancer(gameDispatcherConfig.LoadBalancer)
			}

			gameIDStr := instance.Metadata["clusterId"]

			gameID, err := strconv.ParseUint(gameIDStr, 10, 8)
			if err != nil {
				panic("gameDispatcherConfig dispatcher instance clusterId is not int")
			}

			gameDispatcher := gameDispatcherMap[game][uint8(gameID)]
			if gameDispatcher != nil {
				continue
			}

			gameDispatcher = NewGameDispatcher(game, uint8(gameID), instance.Ip, instance.Port)

			for i := uint8(0); i < channelNum; i++ {
				gameDispatcher.channels[i] = NewDispatcherChannel(i, gameDispatcher)
			}

			gameDispatcherMap[game][uint8(gameID)] = gameDispatcher

			gameDispatcher.Run()
		}
	}
}

func SetDispatcherClientPacketQueue(queue chan *pktconn.Packet) {
	dispatcherClientPacketQueue = queue
}

func ChooseGameDispatcher(game string, entityID int64) *GameDispatcher {
	loadBalancer := gameLoadBalancerMap[game]
	if loadBalancer == nil {
		log.Warnf("game:< %s > loadBalancer is nil", game)
		return nil
	}
	return loadBalancer.Choose(game, entityID)
}

func GetGameDispatcher(game string, gameID uint8) *GameDispatcher {
	loadBalancer := gameLoadBalancerMap[game]
	if loadBalancer == nil {
		log.Warnf("game:< %s > loadBalancer is nil", game)
		return nil
	}
	return loadBalancer.FixedChoose(game, gameID)
}

type LoadBalancer interface {
	Choose(game string, entityID int64) *GameDispatcher
	FixedChoose(game string, gameID uint8) *GameDispatcher
}

func CreateLoadBalancer(loadBalancerType string) LoadBalancer {
	switch loadBalancerType {
	case "polling":
		return NewPollingLoadBalancer()
	default:
		return nil
	}
}

type PollingLoadBalancer struct {
	pollingIndex int8
	gameIDs      []uint8
}

func NewPollingLoadBalancer() *PollingLoadBalancer {
	return &PollingLoadBalancer{}
}

func (l *PollingLoadBalancer) Choose(game string, entityID int64) *GameDispatcher {
	gameDispatchers := gameDispatcherMap[game]
	if gameDispatchers == nil || len(gameDispatchers) == 0 {
		return nil
	}

	if len(gameDispatchers) == len(l.gameIDs) {
		l.pollingIndex++
		if l.pollingIndex >= int8(len(l.gameIDs)) {
			l.pollingIndex = 0
		}
		return gameDispatchers[l.gameIDs[l.pollingIndex]]
	}

	gameIDs := make([]uint8, 0, len(gameDispatchers))
	for gameID := range gameDispatchers {
		gameIDs = append(gameIDs, gameID)
	}
	l.gameIDs = gameIDs

	return gameDispatchers[l.gameIDs[l.pollingIndex]]
}

func (l *PollingLoadBalancer) FixedChoose(game string, gameID uint8) *GameDispatcher {
	gameDispatchers := gameDispatcherMap[game]
	if gameDispatchers == nil || len(gameDispatchers) == 0 {
		return nil
	}

	return gameDispatchers[gameID]
}

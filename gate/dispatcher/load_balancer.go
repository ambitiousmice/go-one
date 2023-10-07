package dispatcher

import (
	"go-one/common/log"
	"reflect"
	"sync/atomic"
)

var loadBalancerTypes = make(map[string]reflect.Type)

func init() {
	AddLoadBalancer("polling", NewPollingLoadBalancer())
	AddLoadBalancer("hash", NewHashLoadBalancer())
}

type LoadBalancer interface {
	Choose(game string, param any) *GameDispatcher
	FixedChoose(game string, gameID uint8) *GameDispatcher
	Init()
}

func AddLoadBalancer(loadBalancerType string, loadBalancer LoadBalancer) {
	if loadBalancerTypes[loadBalancerType] != nil {
		panic("loadBalancer type already registered, loadBalancerType:" + loadBalancerType)
	}

	objVal := reflect.ValueOf(loadBalancer)
	objType := objVal.Type()

	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	loadBalancerTypes[loadBalancerType] = objType

	log.Infof("register dispatcher load balancer: %s", loadBalancerType)

}

func CreateLoadBalancer(loadBalancerType string) LoadBalancer {
	loadBalancerObjType := loadBalancerTypes[loadBalancerType]
	if loadBalancerObjType == nil {
		log.Errorf("dispatcher loadBalancer type not found, loadBalancerType:%s", loadBalancerType)
		return nil
	}

	loadBalancerObj := reflect.New(loadBalancerObjType)
	loadBalancer := loadBalancerObj.Interface().(LoadBalancer)
	loadBalancer.Init()

	return loadBalancer
}

type PollingLoadBalancer struct {
	pollingIndex uint64
	gameIDs      []uint8
}

func NewPollingLoadBalancer() *PollingLoadBalancer {
	return &PollingLoadBalancer{}
}

func (l *PollingLoadBalancer) Choose(game string, param any) *GameDispatcher {
	gameDispatchers := gameDispatcherMap[game]
	if gameDispatchers == nil || len(gameDispatchers) == 0 {
		return nil
	}

	if len(gameDispatchers) == len(l.gameIDs) {
		pollingIndex := uint8(atomic.AddUint64(&l.pollingIndex, 1) % uint64(len(l.gameIDs)))

		return gameDispatchers[l.gameIDs[pollingIndex]]
	}

	gameIDs := make([]uint8, 0, len(gameDispatchers))
	for gameID := range gameDispatchers {
		gameIDs = append(gameIDs, gameID)
	}
	l.gameIDs = gameIDs

	pollingIndex := uint8(atomic.AddUint64(&l.pollingIndex, 1) % uint64(len(l.gameIDs)))

	return gameDispatchers[l.gameIDs[pollingIndex]]
}

func (l *PollingLoadBalancer) FixedChoose(game string, gameID uint8) *GameDispatcher {
	gameDispatchers := gameDispatcherMap[game]
	if gameDispatchers == nil || len(gameDispatchers) == 0 {
		return nil
	}

	return gameDispatchers[gameID]
}

func (l *PollingLoadBalancer) Init() {

}

type HashLoadBalancer struct {
	gameIDs []uint8
}

func NewHashLoadBalancer() *HashLoadBalancer {
	return &HashLoadBalancer{}
}

func (l *HashLoadBalancer) Choose(game string, entityID any) *GameDispatcher {
	gameDispatchers := gameDispatcherMap[game]
	if gameDispatchers == nil || len(gameDispatchers) == 0 {
		return nil
	}

	if len(gameDispatchers) == len(l.gameIDs) {
		index := entityID.(uint64) % uint64(len(l.gameIDs))

		return gameDispatchers[l.gameIDs[index]]
	}

	gameIDs := make([]uint8, 0, len(gameDispatchers))
	for gameID := range gameDispatchers {
		gameIDs = append(gameIDs, gameID)
	}
	l.gameIDs = gameIDs

	index := entityID.(uint64) % uint64(len(l.gameIDs))

	return gameDispatchers[l.gameIDs[index]]
}

func (l *HashLoadBalancer) FixedChoose(game string, gameID uint8) *GameDispatcher {
	gameDispatchers := gameDispatcherMap[game]
	if gameDispatchers == nil || len(gameDispatchers) == 0 {
		return nil
	}

	return gameDispatchers[gameID]
}

func (l *HashLoadBalancer) Init() {

}

package entity

import (
	"github.com/ambitiousmice/go-one/common/log"
	"reflect"
	"sync"
)

var playerMap = map[int64]*Player{}
var playerMutex sync.RWMutex
var playerType reflect.Type

func SetPlayerType(iPlayer IPlayer) {
	objVal := reflect.ValueOf(iPlayer)
	playerType = objVal.Type()

	if playerType.Kind() == reflect.Ptr {
		playerType = playerType.Elem()
	}
}

func GetPlayer(entityID int64) *Player {
	playerMutex.RLock()
	defer playerMutex.RUnlock()
	return playerMap[entityID]
}

func AddPlayer(entityID int64, region int32, gateClusterID uint8) *Player {
	log.Infof("添加用户:%d", entityID)
	player := GetPlayer(entityID)
	if player != nil {
		return player
	}

	iPlayerValue := reflect.New(playerType)
	iPlayer := iPlayerValue.Interface().(IPlayer)
	player = reflect.Indirect(iPlayerValue).FieldByName("Player").Addr().Interface().(*Player)
	player.I = iPlayer
	player.init(entityID, region, gateClusterID)

	playerMutex.Lock()
	defer playerMutex.Unlock()
	playerMap[player.EntityID] = player

	return player
}

func RemovePlayer(entityID int64) {
	player := GetPlayer(entityID)
	if player != nil && player.I != nil {
		player.Destroy()
		player.I.OnDestroy()
	} else {
		log.Infof("移除用户:%d,不存在", entityID)
		return
	}
	playerMutex.Lock()
	defer playerMutex.Unlock()
	delete(playerMap, entityID)
	log.Infof("移除用户:%d", entityID)
}

func GetPlayerCount() int {
	return len(playerMap)
}

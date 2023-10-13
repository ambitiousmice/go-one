package player

import (
	"go-one/common/log"
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

func AddPlayer(entityID int64, gateClusterID uint8) *Player {
	log.Infof("添加用户:%d", entityID)
	player := GetPlayer(entityID)
	if player != nil {
		return player
	}

	iPlayerValue := reflect.New(playerType)
	iPlayer := iPlayerValue.Interface().(IPlayer)
	player = reflect.Indirect(iPlayerValue).FieldByName("Player").Addr().Interface().(*Player)
	player.I = iPlayer
	player.init(entityID, gateClusterID)

	playerMutex.Lock()
	defer playerMutex.Unlock()
	playerMap[player.EntityID] = player

	return player
}

func RemovePlayer(entityID int64) {
	log.Infof("删除用户:%d", entityID)

	player := GetPlayer(entityID)
	if player != nil && player.I != nil {
		player.I.OnClientDisconnected()
	} else {
		log.Infof("删除用户:%d,不存在", entityID)
	}
	playerMutex.Lock()
	defer playerMutex.Unlock()
	delete(playerMap, entityID)
}

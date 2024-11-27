package entity

import (
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/utils"
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
		log.Infof("注意已有用户！！！！！！！！！！:%d", entityID)
		return player
	}

	log.Infof("开始反射获取用户对象:%d", entityID)
	iPlayerValue := reflect.New(playerType)
	iPlayer := iPlayerValue.Interface().(IPlayer)
	player = reflect.Indirect(iPlayerValue).FieldByName("Player").Addr().Interface().(*Player)
	player.I = iPlayer
	log.Infof("开始初始化用户对象:%d", entityID)
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
	for i, _ := range playerMap {
		log.Infof("player_id:" + utils.ToString(i))
	}
	return len(playerMap)
}

package game

import "sync"

var playerMap = map[int64]*BasePlayer{}
var playerMutex sync.RWMutex

func GetPlayer(entityID int64) *BasePlayer {
	playerMutex.RLock()
	defer playerMutex.RUnlock()
	return playerMap[entityID]
}

func AddPlayer(basePlayer *BasePlayer) {
	oldPlayer := GetPlayer(basePlayer.entityID)
	if oldPlayer != nil && oldPlayer.I != nil {
		oldPlayer.I.OnDestroy()
	}
	playerMutex.Lock()
	defer playerMutex.Unlock()
	playerMap[basePlayer.entityID] = basePlayer
}

func RemovePlayer(entityID int64) {
	player := GetPlayer(entityID)
	if player != nil && player.I != nil {
		player.I.OnClientDisconnected()
	}
	playerMutex.Lock()
	defer playerMutex.Unlock()
	delete(playerMap, entityID)
}

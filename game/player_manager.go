package game

var playerMap = map[int64]*BasePlayer{}

func GetPlayer(entityID int64) *BasePlayer {
	return playerMap[entityID]
}

func AddPlayer(basePlayer *BasePlayer) {
	oldPlayer := playerMap[basePlayer.entityID]
	if oldPlayer != nil && oldPlayer.I != nil {
		oldPlayer.I.OnDestroy()
	}
	playerMap[basePlayer.entityID] = basePlayer
}

func removePlayer(basePlayer *BasePlayer) {
	player := playerMap[basePlayer.entityID]
	if player != nil && player.I != nil {
		player.I.OnClientDisconnected()
	}
	delete(playerMap, basePlayer.entityID)
}

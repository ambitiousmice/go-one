package game

// scene type

const (
	SceneTypeLobby = "lobby"
)

// scene match strategy

const (
	SceneStrategyOrder    = "order"
	SceneStrategyRandom   = "random"
	SceneStrategyBalanced = "balanced"
)

// game error msg

const (
	ServerIsFull = "服务器爆满,请稍后再试"
)

// player status
const (
	PlayerStatusInit    = 0
	PlayerStatusOnline  = 5
	PlayerStatusOffline = 10
)

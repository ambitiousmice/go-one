package common

type Coord float32
type Yaw float32
type Speed float32

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

// Game error msg

const (
	ServerIsFull    = "服务器爆满,请稍后再试"
	ReconnectFailed = "重连失败..."
)

// player status
const (
	PlayerStatusInit    = 0
	PlayerStatusOnline  = 5
	PlayerStatusOffline = 10
)

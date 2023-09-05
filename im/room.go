package im

import (
	"go-one/game"
	"sync"
)

type Room struct {
	ID      int32
	Name    string
	Type    int8
	players map[int64]*game.BasePlayer
	mutex   sync.RWMutex
}

package scene_center

import (
	"go-one/game/player"
)

type IScene interface {
	GetSceneType() string
	OnCreated()
	OnDestroyed()
	OnJoined(player *player.Player)
	OnLeft(player *player.Player)
}

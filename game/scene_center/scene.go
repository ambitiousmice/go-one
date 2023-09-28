package scene_center

import (
	"fmt"
	"go-one/game/player"
)

type Scene struct {
	*BaseScene

	I IScene
}

func (r *Scene) init(id int64, sceneType string, maxPlayerNum int) {
	r.BaseScene = NewBaseScene(id, sceneType, maxPlayerNum)

	r.I.OnCreated()
}

func (r *Scene) String() string {
	return fmt.Sprintf("scene info: type=<%s>, id=<%d>", r.Type, r.ID)
}

func (r *Scene) join(player *player.Player) {
	r.AddPlayer(player)

	player.SceneType = r.Type
	player.SceneID = r.ID

	r.I.OnJoined(player)
}

func (r *Scene) leave(player *player.Player) {
	r.RemovePlayer(player)

	r.I.OnLeft(player)

	player.SceneType = ""
	player.SceneID = 0
}

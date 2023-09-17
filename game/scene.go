package game

import "fmt"

type Scene struct {
	*BaseScene

	I IScene
}

func (r *Scene) init(id int64, sceneType string, maxPlayerNum int) {
	r.BaseScene = NewBaseScene(id, sceneType, maxPlayerNum)

	r.I.OnCreated()
}

func (r *Scene) String() string {
	return fmt.Sprintf("scene info: type=<%s>, id=<%d>, playerCount<%d>", r.Type, r.ID, len(r.players))
}

func (r *Scene) Join(player *Player) {
	r.AddPlayer(player)

	player.Scene = r

	r.I.OnJoined(player)

}

func (r *Scene) Leave(player *Player) {
	r.RemovePlayer(player)

	r.I.OnLeft(player)

	player.Scene = nil
}

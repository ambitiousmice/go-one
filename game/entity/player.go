package entity

import (
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/game/common"
)

type Player struct {
	*BasePlayer

	I IPlayer
}

func (p *Player) init(entityID int64, region int32, gateClusterID uint8) {
	p.BasePlayer = NewBasePlayer(entityID, region, gateClusterID)

	p.I.OnCreated()
}

func (p *Player) JoinScene(sceneType string, sceneID int64) {
	submitSceneTask(func() {
		scene := p.Scene
		if scene != nil {
			if scene.Type == sceneType && scene.ID == sceneID {
				p.ReJoinScene()
				return
			}

			scene.leave(p)
		}

		sceneManager := GetSceneManager(sceneType)

		if sceneID == 0 {
			scene = sceneManager.GetSceneByStrategy()
		} else {
			scene = sceneManager.GetScene(sceneID)
		}

		if scene == nil {
			p.SendCommonErrorMsg(common.ServerIsFull)
			return
		}

		scene.join(p)
	})
}

func (p *Player) ReJoinScene() {
	submitSceneTask(func() {
		scene := p.Scene
		if scene == nil {
			p.SendCommonErrorMsg(common.ReconnectFailed)
			return
		}

		scene.join(p)
	})

}

func (p *Player) LeaveScene() {
	submitSceneTask(func() {
		scene := p.Scene
		if scene == nil {
			return
		}

		scene.leave(p)
	})
}

func (p *Player) Move(moveReq *common_proto.MoveReq) {
	submitSceneTask(func() {
		scene := p.Scene
		if scene == nil {
			return
		}

		p.Position.X = common.Coord(moveReq.X)
		p.Position.Y = common.Coord(moveReq.Y)
		p.Position.Z = common.Coord(moveReq.Z)

		p.Speed = common.Speed(moveReq.Speed)
		p.Yaw = common.Yaw(moveReq.Yaw)
		scene.aoiMgr.Moved(&p.AOI, p.Position.X, p.Position.Z)
	})

}

func (p *Player) Destroy() {
	p.Status = common.PlayerStatusOffline
	p.cronTab.Stop()
}

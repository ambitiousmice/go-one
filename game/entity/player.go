package entity

import (
	"go-one/common/common_proto"
	"go-one/game/common"
)

type Player struct {
	*BasePlayer

	I IPlayer
}

func (p *Player) init(entityID int64, gateClusterID uint8) {
	p.BasePlayer = NewBasePlayer(entityID, gateClusterID)

	p.I.OnCreated()
}

func (p *Player) JoinScene(sceneType string, sceneID int64) {
	submitAOITask(func() {
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
	submitAOITask(func() {
		scene := p.Scene
		if scene == nil {
			p.SendCommonErrorMsg(common.ReconnectFailed)
			return
		}

		scene.join(p)
	})

}

func (p *Player) LeaveScene() {
	submitAOITask(func() {
		scene := p.Scene
		if scene == nil {
			return
		}

		scene.leave(p)
	})
}

func (p *Player) Move(moveReq *common_proto.MoveReq) {
	submitAOITask(func() {
		scene := p.Scene
		if scene == nil {
			return
		}

		p.Position.X = moveReq.X
		p.Position.Y = moveReq.Y
		p.Position.Z = moveReq.Z

		p.Speed = moveReq.Speed
		p.Yaw = moveReq.Yaw
		scene.aoiMgr.Moved(&p.AOI, moveReq.X, moveReq.Z)
	})

}

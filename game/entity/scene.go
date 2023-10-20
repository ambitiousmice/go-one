package entity

import (
	"fmt"
	"go-one/game/aoi"
	"go-one/game/common"
)

type Scene struct {
	*BaseScene

	I IScene
}

func (s *Scene) init(id int64, sceneType string, maxPlayerNum int, enableAOI bool, aoiDistance float32) {
	s.BaseScene = NewBaseScene(id, sceneType, maxPlayerNum)
	if enableAOI {
		s.aoiMgr = aoi.NewXZListAOIManager(common.Coord(aoiDistance))
	}
	s.I.OnCreated()
}

func (s *Scene) String() string {
	return fmt.Sprintf("scene info: type=<%s>, id=<%d>", s.Type, s.ID)
}

func (s *Scene) join(player *Player) {
	exist := s.ContainPlayer(player.EntityID)
	if !exist {
		s.AddPlayer(player)

		player.Scene = s
	}

	if s.aoiMgr != nil {
		player.BasePlayer.Position = s.DefaultPosition
		player.SendCreateEntity(player.BasePlayer)
		if !exist {
			s.aoiMgr.Enter(&player.AOI, s.DefaultPosition.X, s.DefaultPosition.Z)
		}
		for neighbor := range player.InterestedBy {
			player.SendCreateEntity(neighbor)
		}
	}

	s.I.OnJoined(player)
}

func (s *Scene) leave(player *Player) {
	s.I.OnLeft(player)

	if s.aoiMgr != nil {
		s.aoiMgr.Leave(&player.AOI)
	}

	s.RemovePlayer(player)

	player.Scene = nil
}

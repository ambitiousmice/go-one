package entity

import (
	"fmt"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/game/aoi"
	"github.com/ambitiousmice/go-one/game/common"
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

	//TODO  上线 开启
	/*s.AddCronTask("entity count", "@every 5s", func() {
		log.Infof("%s 最大承载人数:%d,当前人数:%d", s, s.MaxPlayerNum, s.GetPlayerCount())
	})*/

	s.StartCron()
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
		player.aoiMutex.RLock()
		for neighbor := range player.InterestedBy {
			player.SendCreateEntity(neighbor)
		}
		player.aoiMutex.RUnlock()
	}

	s.I.OnJoined(player)
}

func (s *Scene) leave(player *Player) {
	log.Infof("OnLeft1")

	s.I.OnLeft(player)

	if s.aoiMgr != nil {
		s.aoiMgr.Leave(&player.AOI)
	}

	s.RemovePlayer(player)

	player.Scene = nil
}

package game_client

import (
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/json"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/game/common"
	"time"
)

type CreateEntityProcessor struct {
}

func (p *CreateEntityProcessor) Process(client *Client, param []byte) {
	createEntity := &common_proto.OnCreateEntity{}
	UnPackMsg(param, createEntity)
	s, _ := json.MarshalToString(createEntity)

	if createEntity.EntityID == client.ID {
		log.Infof("create self entity:%s", s)
		client.Position.X = common.Coord(createEntity.X)
		client.Position.Y = common.Coord(createEntity.Y)
		client.Position.Z = common.Coord(createEntity.Z)
		go func() {
			tick := time.Tick(34 * time.Millisecond)
			for {
				select {
				case <-tick:
					client.Position.X = client.Position.X + 0.0000001
					client.Position.Y = client.Position.Y + 0.0000001
					client.Position.Z = client.Position.Z + 0.0000001
					moveReq := &common_proto.MoveReq{
						X:     float32(client.Position.X),
						Y:     float32(client.Position.Y),
						Z:     float32(client.Position.Z),
						Yaw:   float32(client.Yaw),
						Speed: float32(client.Speed),
					}

					client.SendGameData(common_proto.Move, moveReq)
				}
			}
		}()
	} else {
		log.Infof("create other entity:%s", s)
	}
}

func (p *CreateEntityProcessor) GetCmd() uint16 {
	return common_proto.CreateEntity
}

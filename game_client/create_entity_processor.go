package game_client

import (
	"go-one/common/common_proto"
	"go-one/common/json"
	"go-one/common/log"
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
		client.Position.X = createEntity.X
		client.Position.Y = createEntity.Y
		client.Position.Z = createEntity.Z
		go func() {
			tick := time.Tick(50 * time.Millisecond)
			for {
				select {
				case <-tick:
					client.Position.X = client.Position.X + 0.0000001
					client.Position.Y = client.Position.Y + 0.0000001
					client.Position.Z = client.Position.Z + 0.0000001
					moveReq := &common_proto.MoveReq{
						X:     client.Position.X,
						Y:     client.Position.Y,
						Z:     client.Position.Z,
						Yaw:   client.Yaw,
						Speed: client.Speed,
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

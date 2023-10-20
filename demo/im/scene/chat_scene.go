package scene

import (
	"go-one/common/common_proto"
	"go-one/common/log"
	"go-one/demo/im/chat"
	"go-one/demo/im/common"
	"go-one/game/entity"
)

type ChatScene struct {
	entity.Scene

	RoomManager *chat.ChatRoomManager
}

func (r *ChatScene) GetSceneType() string {
	return common.SceneTypeChat
}

func (r *ChatScene) OnCreated() {
	r.RoomManager = chat.NewChatRoomManager()
}

func (r *ChatScene) OnDestroyed() {
	log.Infof("%s destroyed", r)
}

func (r *ChatScene) OnJoined(p *entity.Player) {
	log.Infof("%s joined %s ", p, r)
	joinSceneResp := &common_proto.JoinSceneResp{
		SceneID:   r.ID,
		SceneType: r.Type,
	}

	p.SendGameData(common_proto.JoinScene, joinSceneResp)
}

func (r *ChatScene) OnLeft(p *entity.Player) {
	r.RemovePlayer(p)
}

package main

import (
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/demo/im/proto"
	"github.com/ambitiousmice/go-one/game_client"
	"time"
)

func main() {
	RegisterProcessors()

	imClient := &ImClient{}
	game_client.NewClientServer(imClient).Run()

	for {
		time.Sleep(time.Second)
	}
}

func RegisterProcessors() {
	game_client.RegisterProcessor(&SubscribeRoomProcessor{})
	game_client.RegisterProcessor(&MessageAckProcessor{})

}

type ImClient struct {
	game_client.Client
}

func (ic *ImClient) OnCreated(client *game_client.Client) {

}

func (ic *ImClient) LoginReqWrapper(client *game_client.Client) *common_proto.LoginReq {
	return &common_proto.LoginReq{
		AccountType: consts.TokenLogin,
		Account:     "account",
		Game:        "im",
		EntityID:    ic.ID,
	}
}

func (ic *ImClient) OnLoginSuccess(client *game_client.Client, resp *common_proto.LoginResp) {
	log.Infof("%d login :%s success", resp.EntityID, resp.Game)
}

func (ic *ImClient) OnJoinScene(client *game_client.Client, joinSceneResp *common_proto.JoinSceneResp) {
	log.Infof("%s join scene:type=<%s>,ID=%d success", ic, joinSceneResp.SceneType, joinSceneResp.SceneID)
	if joinSceneResp.SceneType == "lobby" {
		client.SendGameData(common_proto.JoinScene, &common_proto.JoinSceneReq{
			SceneType: "chat",
		})
	}

	if joinSceneResp.SceneType == "chat" {
		client.SendGameData(proto.SubscribeRoom, &proto.SubscribeRoomReq{
			RoomID: 1001,
		})
	}
}

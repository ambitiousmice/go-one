package main

import (
	"go-one/common/common_proto"
	"go-one/common/consts"
	"go-one/common/log"
	"go-one/game_client"
	"time"
)

func main() {
	RegisterProcessors()

	client := &MultiplayerClient{}
	game_client.SetYamlFile("context_multiplayer_client.yaml")
	game_client.NewClientServer(client).Run()

	for {
		time.Sleep(time.Second)
	}
}

func RegisterProcessors() {

}

type MultiplayerClient struct {
	game_client.Client
}

func (ic *MultiplayerClient) OnCreated(client *game_client.Client) {

}

func (ic *MultiplayerClient) EnterGameParamWrapper(client *game_client.Client) *common_proto.EnterGameReq {
	return &common_proto.EnterGameReq{
		AccountType: consts.TokenLogin,
		Account:     "account",
		Game:        "multiplayer",
	}
}

func (ic *MultiplayerClient) OnEnterGameSuccess(client *game_client.Client, resp *common_proto.EnterGameResp) {
	log.Infof("%d enter game:%s success", resp.EntityID, resp.Game)
}

func (ic *MultiplayerClient) OnJoinScene(client *game_client.Client, joinSceneResp *common_proto.JoinSceneResp) {
	log.Infof("%s join scene:type=<%s>,ID=%d success", ic, joinSceneResp.SceneType, joinSceneResp.SceneID)
	if joinSceneResp.SceneType == "lobby" {
		client.SendGameData(common_proto.JoinScene, &common_proto.JoinSceneReq{
			SceneType: "multiplayer",
		})
	}
}

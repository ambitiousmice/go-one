package main

import (
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/game_client"
)

func main() {
	RegisterProcessors()

	client := &MultiplayerClient{}
	game_client.SetYamlFile("context_multiplayer_client.yaml")
	game_client.NewClientServer(client).Run()

	select {}
}

func RegisterProcessors() {

}

type MultiplayerClient struct {
	game_client.Client
}

func (ic *MultiplayerClient) OnCreated(client *game_client.Client) {

}

func (ic *MultiplayerClient) LoginReqWrapper(client *game_client.Client) *common_proto.LoginReq {
	return &common_proto.LoginReq{
		AccountType: consts.TokenLogin,
		Account:     "account",
		Game:        "multiplayer",
		EntityID:    ic.ID,
		Region:      ic.Region,
	}
}

func (ic *MultiplayerClient) OnLoginSuccess(client *game_client.Client, resp *common_proto.LoginResp) {
	log.Infof("%d login:%s success", resp.EntityID, resp.Game)
}

func (ic *MultiplayerClient) OnJoinScene(client *game_client.Client, joinSceneResp *common_proto.JoinSceneResp) {
	log.Infof("%s join scene:type=<%s>,ID=%d success", ic, joinSceneResp.SceneType, joinSceneResp.SceneID)
	if joinSceneResp.SceneType == "lobby" {
		client.SendGameData(common_proto.JoinScene, &common_proto.JoinSceneReq{
			SceneType: "multiplayer",
		})
	}
}

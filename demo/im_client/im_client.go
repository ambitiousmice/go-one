package main

import (
	"go-one/common/consts"
	"go-one/common/log"
	"go-one/common/proto"
	"go-one/game_client"
	"time"
)

func main() {
	imClient := &ImClient{}
	game_client.NewClientServer(imClient).Run()

	for {
		time.Sleep(time.Second)
	}
}

type ImClient struct {
	game_client.Client
}

func (ic *ImClient) OnCreated(client *game_client.Client) {

}

func (ic *ImClient) EnterGameParamWrapper(client *game_client.Client) *proto.EnterGameReq {
	return &proto.EnterGameReq{
		AccountType: consts.TokenLogin,
		Account:     "account",
		Game:        "im",
	}
}

func (ic *ImClient) OnEnterGameSuccess(client *game_client.Client, resp *proto.EnterGameResp) {

}

func (ic *ImClient) OnJoinScene(client *game_client.Client, joinSceneResp *proto.JoinSceneResp) {
	log.Infof("%s join scene:type=<%s>,ID=%d success", ic, joinSceneResp.SceneType, joinSceneResp.SceneID)
}

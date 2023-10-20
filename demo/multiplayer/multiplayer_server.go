package main

import (
	"flag"
	"go-one/common/context"
	entity2 "go-one/demo/multiplayer/entity"
	"go-one/game"
	"go-one/game/entity"
)

func main() {
	flag.Parse()

	context.Init()

	game.InitConfig()

	RegisterProcessor()

	gameServer := game.NewGameServer()

	entity.SetPlayerType(&entity2.Player{})

	entity.RegisterSceneType(&entity2.MultiplayerScene{})

	gameServer.Run()

}

func RegisterProcessor() {
}

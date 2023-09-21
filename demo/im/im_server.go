package main

import (
	"go-one/common/context"
	"go-one/demo/game/processor"
	"go-one/demo/im/player"
	"go-one/game"
)

func main() {

	context.SetYamlFile("context_im.yaml")
	game.SetYamlFile("im.yaml")

	context.Init()

	game.InitGameConfig()

	RegisterProcessor()

	gameServer := game.NewGameServer()

	game.SetPlayerType(&player.ChatPlayer{})

	gameServer.Run()

}

func RegisterProcessor() {
	game.RegisterProcessor(&processor.Test{})
}

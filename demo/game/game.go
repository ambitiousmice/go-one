package main

import (
	"go-one/common/context"
	"go-one/demo/game/processor"
	"go-one/game"
)

func main() {

	context.SetYamlFile("context_game.yaml")

	context.Init()

	game.InitGameConfig()

	RegisterProcessor()

	gameServer := game.NewGameServer()

	gameServer.Run()

}

func RegisterProcessor() {
	game.RegisterProcessor(&processor.Test{})
}

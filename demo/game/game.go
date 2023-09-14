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

	game.SetPlayerType(&DemoPlayer{})

	gameServer.Run()

}

func RegisterProcessor() {
	game.RegisterProcessor(&processor.Test{})
}

type DemoPlayer struct {
	game.Player
}

func (p *DemoPlayer) OnCreated() {

}
func (p *DemoPlayer) OnDestroy() {

}

func (p *DemoPlayer) OnClientConnected() {

}
func (p *DemoPlayer) OnClientDisconnected() {

}

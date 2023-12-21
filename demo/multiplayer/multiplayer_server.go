package main

import (
	"flag"
	"github.com/ambitiousmice/go-one/common/context"
	entity2 "github.com/ambitiousmice/go-one/demo/multiplayer/entity"
	"github.com/ambitiousmice/go-one/game"
	"github.com/ambitiousmice/go-one/game/entity"
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

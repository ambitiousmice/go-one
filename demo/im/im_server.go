package main

import (
	"flag"
	"go-one/common/context"
	"go-one/common/mq/kafka"
	"go-one/demo/im/chat"
	"go-one/demo/im/common"
	"go-one/demo/im/message_center"
	"go-one/demo/im/processor"
	"go-one/demo/im/scene"
	"go-one/game"
	player2 "go-one/game/player"
	"go-one/game/processor_center"
	"go-one/game/scene_center"
)

func main() {
	flag.Parse()

	kafka.RegisterConsumerHandler(common.KafkaConsumerHandlerNameChat, &message_center.Consumer{})

	context.Init()

	game.InitConfig()

	RegisterProcessor()

	gameServer := game.NewGameServer()

	player2.SetPlayerType(&chat.ChatPlayer{})

	scene_center.RegisterSceneType(&scene.ChatScene{})

	gameServer.Run()

}

func RegisterProcessor() {
	processor_center.GPM.RegisterProcessor(&processor.PushOneMessageProcessor{})
	processor_center.GPM.RegisterProcessor(&processor.PushRoomMessageProcessor{})
	processor_center.GPM.RegisterProcessor(&processor.SubscribeRoomProcessor{})
	processor_center.GPM.RegisterProcessor(&processor.UnsubscribeRoomProcessor{})
}

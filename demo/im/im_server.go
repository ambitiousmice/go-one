package main

import (
	"flag"
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/common/mq/kafka"
	"github.com/ambitiousmice/go-one/demo/im/chat"
	"github.com/ambitiousmice/go-one/demo/im/common"
	"github.com/ambitiousmice/go-one/demo/im/message_center"
	"github.com/ambitiousmice/go-one/demo/im/processor"
	"github.com/ambitiousmice/go-one/demo/im/scene"
	"github.com/ambitiousmice/go-one/game"
	"github.com/ambitiousmice/go-one/game/entity"
	"github.com/ambitiousmice/go-one/game/processor_center"
)

func main() {
	flag.Parse()

	kafka.RegisterConsumerHandler(common.KafkaConsumerHandlerNameChat, &message_center.Consumer{})

	context.Init()

	game.InitConfig()

	RegisterProcessor()

	gameServer := game.NewGameServer()

	entity.SetPlayerType(&chat.ChatPlayer{})

	entity.RegisterSceneType(&scene.ChatScene{})

	gameServer.Run()

}

func RegisterProcessor() {
	processor_center.GPM.RegisterProcessor(&processor.PushOneMessageProcessor{})
	processor_center.GPM.RegisterProcessor(&processor.PushRoomMessageProcessor{})
	processor_center.GPM.RegisterProcessor(&processor.SubscribeRoomProcessor{})
	processor_center.GPM.RegisterProcessor(&processor.UnsubscribeRoomProcessor{})
}

package game_client

import (
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/common/log"
	"reflect"
	"time"
)

type ClientServer struct {
	iClientType reflect.Type
}

func NewClientServer(iClient IClient) *ClientServer {
	objVal := reflect.ValueOf(iClient)
	iClientType := objVal.Type()

	if iClientType.Kind() == reflect.Ptr {
		iClientType = iClientType.Elem()
	}
	return &ClientServer{iClientType: iClientType}
}

func (cs *ClientServer) Run() {

	err := InitConfig()
	if err != nil {
		log.Panic(err)
	}
	err = context.InitIDGenerator(Config.IDGeneratorConfig)
	if err != nil {
		log.Panic("init id generator error:" + err.Error())
	}
	RegisterProcessor(&JoinSceneProcessor{})
	RegisterProcessor(&CreateEntityProcessor{})
	RegisterProcessor(&AOISyncProcessor{})
	RegisterProcessor(&BroadcastProcessor{})

	go RunHttpServer()

	for i := 1; i <= Config.ServerConfig.ClientNum; i++ {
		iClientValue := reflect.New(cs.iClientType)
		iClient := iClientValue.Interface().(IClient)

		client := reflect.Indirect(iClientValue).FieldByName("Client").Addr().Interface().(*Client)
		client.I = iClient
		go client.Init(context.NextEntityID()).Run()
		time.Sleep(10 * time.Millisecond)
	}
}

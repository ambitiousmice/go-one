package game_client

import "reflect"

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
		panic(err)
	}

	RegisterProcessor(&JoinSceneProcessor{})
	RegisterProcessor(&CreateEntityProcessor{})
	RegisterProcessor(&AOISyncProcessor{})

	for i := 1; i <= Config.ServerConfig.ClientNum; i++ {
		iClientValue := reflect.New(cs.iClientType)
		iClient := iClientValue.Interface().(IClient)

		client := reflect.Indirect(iClientValue).FieldByName("Client").Addr().Interface().(*Client)
		client.I = iClient
		go client.Init(int64(i)).Run()
	}
}

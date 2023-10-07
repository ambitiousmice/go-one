package game_client

type ClientServer struct {
	iClient IClient
}

func NewClientServer(iClient IClient) *ClientServer {
	return &ClientServer{iClient: iClient}
}

func (cs *ClientServer) Run() {

	err := InitConfig()
	if err != nil {
		panic(err)
	}

	RegisterProcessor(&JoinSceneProcessor{})

	for i := 0; i < Config.ServerConfig.ClientNum; i++ {
		go NewClient(cs.iClient).Run()
	}
}

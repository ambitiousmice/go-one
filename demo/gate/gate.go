package main

import (
	"flag"
	"go-one/common/context"
	"go-one/gate"
	_ "net/http/pprof"
)

func main() {
	flag.Parse()

	context.Init()

	gate.InitConfig()

	gateServer := gate.NewGateServer()

	if gateServer.NeedLogin {
		gateServer.LoginManager = NewDemoLoginManager(gate.GetGateConfig().Params["loginServerUrl"].(string))
	}

	gateServer.Run()
}

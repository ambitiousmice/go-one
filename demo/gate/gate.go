package main

import (
	"flag"
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/gate"
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

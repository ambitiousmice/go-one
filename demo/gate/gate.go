package main

import (
	"go-one/common/context"
	"go-one/gate"
	"time"
)

func main() {

	context.SetYamlFile("context_gate.yaml")
	context.Init()

	gate.SetYamlFile("context_gate.yaml")
	gate.InitConfig()

	gateServer := gate.NewGateServer()

	if gateServer.NeedLogin {
		gateServer.LoginManager = NewDemoLoginManager(gate.GetGateConfig().Params["loginServerUrl"].(string))
	}

	gateServer.Run()

	for {
		time.Sleep(1 * time.Second)
	}
}

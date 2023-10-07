package main

import "go-one/common/context"

func Run() {
	context.SetYamlFile("context_monitor.yaml")

	context.Init()
}

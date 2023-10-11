package main

import (
	"flag"
	"go-one/common/context"
	"go-one/common/utils"
	"go-one/monitor/gate_manager"
	"time"
)
import "github.com/gin-gonic/gin"
import "github.com/gin-contrib/timeout"

func main() {
	flag.Parse()

	context.Init()

	router := gin.Default()

	gin.SetMode(gin.ReleaseMode)

	router.Use(timeoutMiddleware())

	gateGroup := router.Group("/gate")
	{
		gateGroup.GET("/choose/:partition", gate_manager.ChooseGate)
		gateGroup.GET("/choose/test", gate_manager.ChooseGateTest)
		gateGroup.GET("/collectData", gate_manager.CollectData)
	}

	addr := ":" + utils.ToString(context.GetOneConfig().Nacos.Instance.Port)
	router.Run(addr)
}

func timeoutMiddleware() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(3000*time.Millisecond),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
	)
}

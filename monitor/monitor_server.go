package main

import (
	"flag"
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/utils"
	"github.com/ambitiousmice/go-one/monitor/config"
	"github.com/ambitiousmice/go-one/monitor/gate_manager"
	"time"
)
import "github.com/gin-gonic/gin"
import "github.com/gin-contrib/timeout"

func main() {
	flag.Parse()

	context.Init()

	config.InitConfig()

	router := gin.Default()

	gin.SetMode(gin.ReleaseMode)

	router.Use(timeoutMiddleware())

	appGroup := router.Group("/app-api/monitor/gate")
	{
		appGroup.GET("/choose/:partition", gate_manager.ChooseGate)
		appGroup.GET("/choose/inner", gate_manager.ChooseGateInner)
		appGroup.GET("/choose/test", gate_manager.ChooseGateTest)
	}

	gateGroup := router.Group("/gate")
	{
		gateGroup.GET("/collectData", gate_manager.CollectData)
		gateGroup.POST("/broadcast", gate_manager.Broadcast)
	}

	addr := ":" + utils.ToString(context.GetOneConfig().Nacos.Instance.Port)
	log.Infof("server run with:%s", addr)
	err := router.Run(addr)
	if err != nil {
		log.Errorf("server run with error:%s", err)
		return
	}

}

func timeoutMiddleware() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(3000*time.Millisecond),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
	)
}

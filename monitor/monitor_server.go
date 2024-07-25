package main

import (
	"flag"
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/utils"
	"github.com/ambitiousmice/go-one/monitor/config"
	"github.com/ambitiousmice/go-one/monitor/server_manager"
	"time"
)
import "github.com/gin-gonic/gin"
import "github.com/gin-contrib/timeout"

func main() {
	flag.Parse()

	context.Init()

	config.InitConfig()

	server_manager.Init()

	router := gin.Default()

	gin.SetMode(gin.ReleaseMode)

	router.Use(timeoutMiddleware())

	appGroup := router.Group("/app-api/monitor/gate")
	{
		appGroup.GET("/choose/:partition", server_manager.ChooseGate)
		appGroup.GET("/choose/inner", server_manager.ChooseGateInner)
		appGroup.GET("/choose/test", server_manager.ChooseGateTest)
	}

	serverGroup := router.Group("/app-api/monitor/server")
	{
		serverGroup.POST("/queryServerInfo", server_manager.QueryServerInfo)
		serverGroup.POST("/queryServerInfos", server_manager.QueryServerInfos)
	}

	gateGroup := router.Group("/gate")
	{
		gateGroup.POST("/collectData", server_manager.CollectData)
		gateGroup.POST("/broadcast", server_manager.Broadcast)
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

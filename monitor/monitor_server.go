package main

import (
	"go-one/common/context"
	"go-one/common/json"
	"go-one/monitor/gate_manager"
	"net/http"
	"strconv"
)
import "github.com/gin-gonic/gin"

func main() {
	context.SetYamlFile("context_monitor.yaml")

	context.Init()

	router := gin.Default()
	gateGroup := router.Group("/gate")
	{
		gateGroup.GET("/choose/:partition/", chooseGate)
	}

	router.Run(":8080")
}

type UserInfo struct {
	UserId int64
}

func chooseGate(c *gin.Context) {
	userInfoStr := c.Request.Header.Get("user_info")
	userInfo := &UserInfo{}
	err := json.UnmarshalFromString(userInfoStr, userInfo)
	if err != nil {
		return
	}
	partitionStr := c.Param("partition")
	partition, err := strconv.Atoi(partitionStr)
	gateInfo := gate_manager.GetGateInfo(userInfo.UserId, int64(partition))
	c.JSON(http.StatusOK, gin.H{
		"code": "0",
		"data": gateInfo,
		"msg":  "success",
	})
}

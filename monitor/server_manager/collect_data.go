package server_manager

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type ServerDataReq struct {
	ServerName      string
	GroupID         int64
	ClusterId       int64
	ConnectionCount int64
	TotalMemory     float64
	UsageMemory     float64
	Metadata        map[string]string
}

func CollectData(c *gin.Context) {
	var req ServerDataReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	serverInfo := &ServerInfo{
		ServerName:            req.ServerName,
		GroupID:               req.GroupID,
		ClusterId:             req.ClusterId,
		ConnectionCount:       req.ConnectionCount,
		TotalMemory:           req.TotalMemory,
		UsageMemory:           req.UsageMemory,
		LastCommunicationTime: time.Now().UnixMilli(),
		Metadata:              req.Metadata,
	}

	AddServer(serverInfo)

	c.JSON(http.StatusOK, gin.H{
		"code": "0",
		"msg":  "success",
	})
}

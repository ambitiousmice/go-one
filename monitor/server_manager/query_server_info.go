package server_manager

import (
	"github.com/ambitiousmice/go-one/common/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

type QueryServerInfoReq struct {
	ServerName string
	GroupID    int64
	ClusterId  int64
}

type QueryServerInfosReq struct {
	ServerInfos []QueryServerInfoReq
}

func QueryServerInfo(c *gin.Context) {
	var req QueryServerInfoReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	key := req.ServerName + "_" + utils.ToString(req.GroupID) + "_" + utils.ToString(req.ClusterId)
	serverInfo := serverContext[key]

	c.JSON(http.StatusOK, gin.H{
		"code": "0",
		"msg":  serverInfo,
	})
}

func QueryServerInfos(c *gin.Context) {
	var req QueryServerInfosReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serverInfos := make([]ServerInfo, 0)
	for _, info := range req.ServerInfos {
		key := info.ServerName + "_" + utils.ToString(info.GroupID) + "_" + utils.ToString(info.ClusterId)
		serverInfo := serverContext[key]
		if serverInfo != nil {
			serverInfos = append(serverInfos, *serverInfo)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "0",
		"msg":  serverInfos,
	})
}

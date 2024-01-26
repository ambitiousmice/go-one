package gate_manager

import (
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/monitor/util/gin_util"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CollectData(c *gin.Context) {
	partition, err := gin_util.GetQueryInt(c, consts.GroupId)
	if err != nil {
		return
	}
	clusterID, err := gin_util.GetQueryInt(c, consts.ClusterId)
	if err != nil {
		return
	}
	connectionCount, err := gin_util.GetQueryInt(c, consts.ConnectionCount)
	if err != nil {
		return
	}

	gateInfo := GetGateInfo(int64(partition), int64(clusterID))
	if gateInfo != nil {
		gateInfo.ConnectionCount = int64(connectionCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "0",
		"msg":  "success",
	})
}

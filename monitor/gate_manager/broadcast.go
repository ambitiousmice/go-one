package gate_manager

import (
	"github.com/gin-gonic/gin"
	"go-one/common/common_proto"
	"go-one/common/consts"
	"go-one/common/json"
	"go-one/common/mq/kafka"
	"go-one/common/utils"
	"net/http"
)

type BroadcastReq struct {
	Type string
	Data any
}

func Broadcast(c *gin.Context) {
	var req BroadcastReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bytes, err := json.Marshal(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := &common_proto.GateBroadcastMsg{
		Type: req.Type,
		Data: bytes,
	}
	kafka.Producer.SendMessage(consts.GateBroadcastTopic, utils.ToString(req.Type), msg)

	c.JSON(http.StatusOK, gin.H{
		"code": "0",
		"msg":  "success",
	})
}

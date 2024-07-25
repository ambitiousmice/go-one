package server_manager

import (
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/json"
	"github.com/ambitiousmice/go-one/common/mq/kafka"
	"github.com/ambitiousmice/go-one/common/utils"
	"github.com/gin-gonic/gin"
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

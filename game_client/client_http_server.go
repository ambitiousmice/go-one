package game_client

import (
	"flag"
	"github.com/ambitiousmice/go-one/common/json"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/golang/protobuf/proto"
	"net/http"
	"reflect"
	"time"
)
import "github.com/gin-gonic/gin"
import "github.com/gin-contrib/timeout"

func RunHttpServer() {
	flag.Parse()

	router := gin.Default()

	gin.SetMode(gin.ReleaseMode)

	router.Use(timeoutMiddleware())

	gateGroup := router.Group("/test")
	{
		gateGroup.POST("/cmd", httpHandler)
	}

	addr := ":18888"
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

var cmdParamContext = make(map[uint16]reflect.Type)

func RegisterCmdParam(cmd uint16, param any) {
	objVal := reflect.ValueOf(param)
	paramType := objVal.Type()

	if paramType.Kind() == reflect.Ptr {
		paramType = paramType.Elem()
	}
	cmdParamContext[cmd] = paramType
}

type TestCmdReq struct {
	PID  int64  `json:"pid"`
	Cmd  uint16 `json:"cmd"`
	Data any    `json:"data"`
}

func httpHandler(c *gin.Context) {
	var req TestCmdReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if cmdParamContext[req.Cmd] != nil {
		reqData := reflect.New(cmdParamContext[req.Cmd]).Interface().(proto.Message)

		dataBytes, err := json.Marshal(req.Data)

		err = json.Unmarshal(dataBytes, &reqData)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ClientContext[req.PID].SendGameData(req.Cmd, reqData)
	} else {
		ClientContext[req.PID].SendGameData(req.Cmd, nil)

	}

	c.JSON(http.StatusOK, gin.H{
		"code": "0",
		"msg":  "success",
	})
}

package game_client

import (
	"flag"
	"github.com/ambitiousmice/go-one/common/json"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/game/common"
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

var cmdParamNames = make(map[uint16]string)

var respParamContext = make(map[uint16]reflect.Type)

func RegisterCmdParam(cmd uint16, param any) {
	objVal := reflect.ValueOf(param)
	paramType := objVal.Type()

	if paramType.Kind() == reflect.Ptr {
		paramType = paramType.Elem()
	}
	cmdParamContext[cmd] = paramType
}

func RegisterCmdParamAndResp(cmdName string, cmd uint16, param any, resp any) {

	if cmdName != "" {
		cmdParamNames[cmd] = cmdName
	}

	if param != nil {
		objVal := reflect.ValueOf(param)
		paramType := objVal.Type()

		if paramType.Kind() == reflect.Ptr {
			paramType = paramType.Elem()
		}
		cmdParamContext[cmd] = paramType
	}

	if resp != nil {
		objVal := reflect.ValueOf(resp)
		paramType := objVal.Type()

		if paramType.Kind() == reflect.Ptr {
			paramType = paramType.Elem()
		}
		respParamContext[cmd] = paramType
	}

}

func DefaultProcessor(client *Client, Cmd uint16, param []byte) {
	log.Infof("%s: serviceCmd:%s", cmdParamNames[Cmd], Cmd)

	if respParamContext[Cmd] == nil {
		if param == nil || len(param) == 0 {
			log.Infof("%s: serviceCmd:%s，未返回数据", cmdParamNames[Cmd], Cmd)
			return
		}
		if param != nil && len(param) > 0 {
			log.Infof("未注册返回值类型，serviceCmd:%s，未返回数据长度：%s ,resp: ", Cmd, len(param), string(param))
			return
		}
	}

	if respParamContext[Cmd] != nil {
		respData := reflect.New(respParamContext[Cmd]).Interface().(proto.Message)

		err := common.UnPackMsg(param, respData)
		if err != nil {
			log.Errorf("%s:serviceCmd：%s unpack msg error: %s", cmdParamNames[Cmd], Cmd, err.Error())
			return
		}
		s, _ := json.MarshalToString(respData)
		log.Infof("%s: serviceCmd:%s,返回数据:%s", cmdParamNames[Cmd], Cmd, s)
	}
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

		s, _ := json.MarshalToString(req.Data)
		log.Infof("cmd: %s ,请求参数:%s", req.Cmd, s)

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

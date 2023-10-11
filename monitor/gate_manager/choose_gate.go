package gate_manager

import (
	"github.com/gin-gonic/gin"
	"go-one/common/json"
	"net/http"
	"strconv"
)

type UserInfo struct {
	UserId int64
}

func ChooseGate(c *gin.Context) {
	userInfoStr := c.Request.Header.Get("user_info")
	userInfo := &UserInfo{}
	err := json.UnmarshalFromString(userInfoStr, userInfo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": "1",
			"msg":  "请先登录",
		})
		return
	}
	partitionStr := c.Param("partition")
	partition, err := strconv.Atoi(partitionStr)

	gateInfo := ChooseGateInfo(int64(partition), userInfo.UserId)
	resp := &chooseGateResp{
		WsAddr:  gateInfo.WsAddr,
		TcpAddr: gateInfo.TcpAddr,
		Version: gateInfo.Version,
	}
	c.JSON(http.StatusOK, gin.H{
		"code": "0",
		"data": resp,
		"msg":  "success",
	})
}

type chooseGateResp struct {
	WsAddr  string
	TcpAddr string
	Version string
}

func ChooseGateTest(c *gin.Context) {
	partitionStr := c.Query("partition")
	partition, _ := strconv.Atoi(partitionStr)

	userIdStr := c.Query("entityID")
	userId, _ := strconv.Atoi(userIdStr)

	gateInfo := ChooseGateInfo(int64(partition), int64(userId))
	resp := &chooseGateResp{
		WsAddr:  gateInfo.WsAddr,
		TcpAddr: gateInfo.TcpAddr,
		Version: gateInfo.Version,
	}
	c.JSON(http.StatusOK, gin.H{
		"code": "0",
		"data": resp,
		"msg":  "success",
	})
}

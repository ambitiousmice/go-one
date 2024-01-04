package gate_manager

import (
	"github.com/ambitiousmice/go-one/common/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"strconv"
)

type UserInfo struct {
	UserId   int64 `json:"userId"`
	PlayerId int64 `json:"playerId"`
	RegionId int64 `json:"regionId"`
}

func ChooseGate(c *gin.Context) {
	userInfoStr := c.Request.Header.Get("user_info")
	userInfoStr, err := url.QueryUnescape(userInfoStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": "1",
			"msg":  "请先登录",
		})
		return
	}

	userInfo := &UserInfo{}
	err = json.UnmarshalFromString(userInfoStr, userInfo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": "1",
			"msg":  "请先登录",
		})
		return
	}
	partitionStr := c.Param("partition")
	partition, err := strconv.Atoi(partitionStr)

	gateInfo := ChooseGateInfo(int64(partition), userInfo.PlayerId)
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

func ChooseGateInner(c *gin.Context) {
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

type chooseGateResp struct {
	WsAddr  string `json:"wsAddr"`
	TcpAddr string `json:"tcpAddr"`
	Version string `json:"version"`
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

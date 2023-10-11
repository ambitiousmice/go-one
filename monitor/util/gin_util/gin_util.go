package gin_util

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetQueryInt(c *gin.Context, key string) (int, error) {
	valueStr := c.Query(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": "1",
			"msg":  "参数异常",
		})
		return value, err
	}
	return value, err
}

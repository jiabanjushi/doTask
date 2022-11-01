package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"net/http"
)

func ReturnDataLIst2000(context *gin.Context, data interface{}, count int) {
	context.JSON(http.StatusOK, gin.H{
		"code":   client.SuccessReturnCode,
		"result": data,
		"count":  count,
	})
}

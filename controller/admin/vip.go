package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
)

func OperationVip(c *gin.Context) {
	action := c.Query("action")
	if action == "select" {
		v := make([]model.Vip, 0)
		mysql.DB.Find(&v)
		client.ReturnSuccess2000DataCode(c, v, "ok")
		return
	}

}

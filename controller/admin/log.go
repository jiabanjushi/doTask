package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"strconv"
)

// OperationLogger 日志
func OperationLogger(c *gin.Context) {

	action := c.Query("action")
	if action == "select" {

		//普通查询
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.Log, 0)
		db := mysql.DB.Where("kinds=?", c.PostForm("kinds"))

		//查询条件

		if content, isE := c.GetPostForm("content"); isE == true {
			db = db.Where("content  like ?", "%"+content+"%")
		}

		var total int
		//条件
		db.Model(model.Log{}).Count(&total)
		db = db.Model(&model.Log{}).Offset((page - 1) * limit).Limit(limit).Order("created desc")
		db.Find(&sl)
		ReturnDataLIst2000(c, sl, total)

	}
}

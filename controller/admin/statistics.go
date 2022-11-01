package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"strconv"
	"strings"
)

// OperationFirstPage 数据首页操作
func OperationFirstPage(c *gin.Context) {
	action := c.Query("action")
	who, _ := c.Get("who")
	whoMap := who.(model.Admin)
	//ctName, _ := mmdb.GetCountryForIp(c.ClientIP())
	if action == "select" {
		operation := c.PostForm("operation")
		//首页数据
		if operation == "firstPage" {

		}

		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.Statistics, 0)
		db := mysql.DB
		var total int
		if whoMap.AgencyUsername != "" {
			var p []int
			arrayUg := strings.Split(whoMap.AgencyUsername, ",")
			for _, s := range arrayUg {
				us := make([]model.User, 0)
				mysql.DB.Where("top_agent=?", s).Find(&us)
				for _, u := range us {
					p = append(p, u.ID)
				}
			}
			db = db.Where("user_id  in (?)", p)
		}

		db = db.Where("kinds=?", 1)
		db.Model(model.Statistics{}).Count(&total)
		db = db.Model(&model.Statistics{}).Offset((page - 1) * limit).Limit(limit).Order("updated desc")
		db.Find(&sl)

		ReturnDataLIst2000(c, sl, total)

	}
}

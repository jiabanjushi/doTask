package client

import (
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
)

// GetSlideshow GeeSlideshow 获取轮播图
func GetSlideshow(c *gin.Context) {
	countryN := c.PostForm("country_name")
	co := model.Country{}
	err := mysql.DB.Where("country_name=?", countryN).First(&co).Error
	if err != nil {
		ReturnErr101Code(c, map[string]interface{}{"identification": "NoThisCountry", "msg": NoThisCountry})
		return
	}
	sl := make([]model.Slideshow, 0)
	mysql.DB.Where("status=?", 1).Find(&sl)
	ReturnSuccess2000DataCode(c, sl, "OK")
	return

}

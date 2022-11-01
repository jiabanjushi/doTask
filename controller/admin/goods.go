package admin

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mmdb"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"strconv"
	"time"
)

// OperationGoods 商品图片
func OperationGoods(c *gin.Context) {
	action := c.Query("action")
	who, _ := c.Get("who")
	whoMap := who.(model.Admin)
	ctName, _ := mmdb.GetCountryForIp(c.ClientIP())
	if action == "add" {
		gN := c.PostForm("goods_name")
		gI, _ := c.FormFile("goods_images")
		//判断是否重复上传
		err := mysql.DB.Where("goods_name=?", gN).First(&model.Goods{}).Error
		if err == nil {
			client.ReturnErr101Code(c, "不要重复上传")
			return
		}
		filepath := "static/goods/" + gI.Filename + ".png"
		err = c.SaveUploadedFile(gI, filepath)
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		mysql.DB.Save(&model.Goods{GoodsName: gN, GoodsImages: filepath, Created: time.Now().Unix()})
		client.ReturnSuccess2000Code(c, "上传成功")
		marshal, err := json.Marshal(model.Goods{GoodsName: gN, GoodsImages: filepath, Created: time.Now().Unix()})
		if err == nil {
			LO := model.Log{Country: ctName, Ip: c.ClientIP(), Content: whoMap.AdminUser + "|" + "任务图片,添加新的数据:" + string(marshal), Kinds: 3}
			LO.CreateLogger(mysql.DB)
		}
		return
	}

	if action == "select" {
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.Goods, 0)
		db := mysql.DB
		var total int
		db.Model(model.Goods{}).Count(&total)
		db = db.Model(&model.Goods{}).Offset((page - 1) * limit).Limit(limit).Order("created desc")
		db.Find(&sl)
		ReturnDataLIst2000(c, sl, total)
	}

	if action == "delete" {
		id := c.PostForm("ID")
		//修改的id是否存在
		sli := model.Goods{}
		err := mysql.DB.Where("id=?", id).First(&sli).Error
		if err != nil {
			client.ReturnErr101Code(c, "删除的任务图片已经不存在了")
			return
		}
		err = mysql.DB.Model(&model.Goods{}).Where("id=?", id).Delete(&model.Goods{}).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "删除成功")
		LO := model.Log{Country: ctName, Ip: c.ClientIP(), Content: whoMap.AdminUser + "|" + "删除任务图片,id:" + id, Kinds: 3}
		LO.CreateLogger(mysql.DB)
		return
	}

}

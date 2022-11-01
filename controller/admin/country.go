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

func OperationCountry(c *gin.Context) {
	action := c.Query("action")

	who, _ := c.Get("who")
	whoMap := who.(model.Admin)
	ctName, _ := mmdb.GetCountryForIp(c.ClientIP())

	if action == "select" {
		cc := make([]model.Country, 0)
		db := mysql.DB
		if Status, isExist := c.GetPostForm("status"); isExist == true {
			st, _ := strconv.Atoi(Status)
			db = db.Where("status=?", st)
			db.Find(&cc)
			client.ReturnSuccess2000DataCode(c, cc, "ok")
			return
		}

		//分水岭
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.Country, 0)
		var total int
		db.Model(model.Country{}).Count(&total)
		db = db.Model(&model.Country{}).Offset((page - 1) * limit).Limit(limit).Order("updated desc")
		db.Find(&sl)
		ReturnDataLIst2000(c, sl, total)
		return
	}
	if action == "update" {

		id := c.PostForm("id")
		err := mysql.DB.Where("id=?", id).First(&model.Country{}).Error
		if err != nil {
			client.ReturnErr101Code(c, "修改的国家不存在")
			return
		}
		db := mysql.DB.Model(&model.Country{}).Where("id=?", id)
		//修改状态(状态是单独修改)
		if Status, isExist := c.GetPostForm("status"); isExist == true {
			st, _ := strconv.Atoi(Status)
			if st < 1 || st > 2 {
				client.ReturnErr101Code(c, "status非法")
				return
			}
			err := db.Update(&model.Country{Updated: time.Now().Unix(), Status: st}).Error
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}
			client.ReturnSuccess2000Code(c, "修改成功")
			LO := model.Log{Country: ctName, Ip: c.ClientIP(), Content: whoMap.AdminUser + "|" + "更新了国家管理id:" + id + ",更新后的结果,status=" + strconv.Itoa(st), Kinds: 3}
			LO.CreateLogger(mysql.DB)
			return
		}
		//大杂烩一起修改
		country := model.Country{Updated: time.Now().Unix(), CountryName: c.PostForm("country_name")}
		err = db.Update(&country).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "修改成功")
		//添加修改日志
		LO := model.Log{Country: ctName, Ip: c.ClientIP(), Content: whoMap.AdminUser + "|" + "更新了国家管理id:" + id + ",更新后的结果,country_name=" + c.PostForm("country_name"), Kinds: 3}
		LO.CreateLogger(mysql.DB)
		return
	}

	if action == "add" {
		var ac AddCountry
		if err := c.ShouldBind(&ac); err != nil {
			client.ReturnVerifyErrCode(c, err)
			return
		}

		newCountry := model.Country{
			CountryName: c.PostForm("country_name"),
			Created:     time.Now().Unix(),
			Updated:     time.Now().Unix(),
			Status:      1,
		}

		//判断是否重复添加
		err := mysql.DB.Where("country_name=?", newCountry.CountryName).First(&model.Country{}).Error
		if err == nil {
			client.ReturnErr101Code(c, "不要重复添加")
			return
		}
		err = mysql.DB.Save(&newCountry).Error
		if err != nil {
			client.ReturnErr101Code(c, err)
			return
		}
		client.ReturnSuccess2000Code(c, "添加成功")
		marshal, err := json.Marshal(newCountry)
		if err == nil {
			LO := model.Log{Country: ctName, Ip: c.ClientIP(), Content: whoMap.AdminUser + "|" + "国家系统,添加新的数据:" + string(marshal), Kinds: 3}
			LO.CreateLogger(mysql.DB)
		}
		return
	}

}

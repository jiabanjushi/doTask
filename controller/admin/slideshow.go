package admin

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mmdb"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"github.com/wangyi/GinTemplate/tools"
	"strconv"
	"strings"
	"time"
)

//对轮播图进行增删改查

func OperationSlideshow(c *gin.Context) {
	action := c.Query("action")
	who, _ := c.Get("who")
	whoMap := who.(model.Admin)
	ctName, _ := mmdb.GetCountryForIp(c.ClientIP())
	if action == "select" {
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.Slideshow, 0)
		db := mysql.DB
		var total int
		db.Model(model.Slideshow{}).Count(&total)
		db = db.Model(&model.Slideshow{}).Offset((page - 1) * limit).Limit(limit).Order("updated desc")
		db.Find(&sl)
		for i, slideshow := range sl {
			country := model.Country{}
			err := mysql.DB.Where("id=?", slideshow.CountryId).First(&country).Error
			if err == nil {
				sl[i].Country = country.CountryName
			}
		}

		ReturnDataLIst2000(c, sl, total)
	}
	if action == "add" {
		file, _ := c.FormFile("file")
		nameArray := strings.Split(file.Filename, ".")
		if len(nameArray) < 1 || tools.IsArray([]string{"jpg", "png"}, nameArray[1]) == false {
			client.ReturnErr101Code(c, "上传文件格式有误")
			return
		}
		//保存文件
		//判断其他参数是否正确   国家?
		countryId, _ := strconv.Atoi(c.PostForm("country_id"))
		err := mysql.DB.Where("id=? and status=?", countryId, 1).First(&model.Country{}).Error
		if err != nil {
			client.ReturnErr101Code(c, "国家非法,或者已经被禁用")
			return
		}

		// "static/upload/"  这个目录程序初始化已经创建了
		filepath := "static/upload/" + file.Filename + ".png"

		//不要去添加同一张图片
		err = mysql.DB.Where("image_url =? and country_id=?", filepath, countryId).First(&model.Slideshow{}).Error
		if err == nil {
			client.ReturnErr101Code(c, "同一个国家只能添加同一张图片")
			return

		}

		err = c.SaveUploadedFile(file, filepath)
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		//入库
		sl := model.Slideshow{
			CountryId: countryId,
			Created:   time.Now().Unix(),
			Updated:   time.Now().Unix(),
			Status:    1,
			Remark:    c.PostForm("remark"),
			ImageUrl:  filepath,
		}

		err = mysql.DB.Save(&sl).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}

		marshal, err := json.Marshal(sl)
		if err == nil {
			LO := model.Log{Country: ctName, Ip: c.ClientIP(), Content: whoMap.AdminUser + "|" + "轮播图系统,添加新的数据:" + string(marshal), Kinds: 3}
			LO.CreateLogger(mysql.DB)
		}

		client.ReturnSuccess2000Code(c, "添加成功")

	}
	if action == "update" {
		id := c.PostForm("id")
		//修改的id是否存在
		sli := model.Slideshow{}
		err := mysql.DB.Where("id=?", id).First(&sli).Error
		if err != nil {
			client.ReturnErr101Code(c, "修改的轮播图已经不存在了")
			return
		}
		db := mysql.DB.Model(&model.Slideshow{}).Where("id=?", id)
		//修改轮播图的状态
		if Status, isExist := c.GetPostForm("status"); isExist == true {
			st, _ := strconv.Atoi(Status)
			if st < 1 || st > 2 {
				client.ReturnErr101Code(c, "status非法")
				return
			}
			err := db.Update(&model.Slideshow{Updated: time.Now().Unix(), Status: st}).Error
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}
			client.ReturnSuccess2000Code(c, "修改成功")
			LO := model.Log{Country: ctName, Ip: c.ClientIP(), Content: whoMap.AdminUser + "|" + "更新了轮播图管理id:" + id + ",更新后的结果,status=" + strconv.Itoa(st), Kinds: 3}
			LO.CreateLogger(mysql.DB)
			return
		}
		//更新国家国家 和备注
		countryId, _ := strconv.Atoi(c.PostForm("country_id"))
		err = mysql.DB.Where("id=? and status=?", countryId, 1).First(&model.Country{}).Error
		if err != nil {
			client.ReturnErr101Code(c, "国家非法,或者已经被禁用")
			return
		}
		newSli := model.Slideshow{
			CountryId: countryId,
			Remark:    c.PostForm("remark"),
		}
		err = db.Update(&newSli).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "修改成功")
		marshal, err := json.Marshal(&newSli)
		if err == nil {
			LO := model.Log{Country: ctName, Ip: c.ClientIP(), Content: whoMap.AdminUser + "|" + "修改轮播图管理,id:" + id + " 修改内容:" + string(marshal), Kinds: 3}
			LO.CreateLogger(mysql.DB)
		}
		return
	}
	if action == "delete" {
		id := c.PostForm("id")
		//修改的id是否存在
		sli := model.Slideshow{}
		err := mysql.DB.Where("id=?", id).First(&sli).Error
		if err != nil {
			client.ReturnErr101Code(c, "删除的轮播图已经不存在了")
			return
		}
		err = mysql.DB.Model(&model.Slideshow{}).Where("id=?", id).Delete(&model.Slideshow{}).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "删除成功")
		LO := model.Log{Country: ctName, Ip: c.ClientIP(), Content: whoMap.AdminUser + "|" + "删除轮播图,id:" + id, Kinds: 3}
		LO.CreateLogger(mysql.DB)
		return
	}
	return
}

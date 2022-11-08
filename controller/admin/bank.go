package admin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"strconv"
	"strings"
	"time"
)

func OperationBank(c *gin.Context) {
	action := c.Query("action")
	if action == "select" {
		operation := c.PostForm("operation")

		if operation == "bank_card" {
			//查询bankCard
			limit, _ := strconv.Atoi(c.PostForm("limit"))
			page, _ := strconv.Atoi(c.PostForm("page"))
			sl := make([]model.BankCard, 0)
			db := mysql.DB

			db = db.Where("bank_pay_id=?", c.PostForm("bank_pay_id"))
			var total int
			//条件
			db.Model(model.BankCard{}).Count(&total)
			db = db.Model(&model.BankCard{}).Offset((page - 1) * limit).Limit(limit).Order("created desc")
			db.Find(&sl)
			ReturnDataLIst2000(c, sl, total)
			return
		}

		if operation == "bank_card_add" {
			data := c.PostForm("data")
			Bi, _ := strconv.Atoi(c.PostForm("bank_pay_id"))
			dataArray := strings.Split(data, "\n")
			if len(dataArray) < 2 {
				client.ReturnErr101Code(c, "长度错误")
				return
			}
			for _, s := range dataArray {
				bc := model.BankCard{}
				bc.Status = 1
				bc.Created = time.Now().Unix()
				bc.BankPayId = Bi
				bc.Updated = time.Now().Unix()
				count := strings.Index(s, " ")
				bc.BankCode = strings.TrimSpace(s[:count])
				bc.BankName = strings.TrimSpace(s[count:len(s)])
				err := mysql.DB.Where("bank_pay_id=? and bank_code=? and bank_name=?", Bi, bc.BankCode, bc.BankName).First(&model.BankCard{}).Error
				if err != nil {
					err := mysql.DB.Save(&bc).Error
					if err != nil {
						fmt.Println(err.Error())
					}

				}

			}

			client.ReturnSuccess2000Code(c, "OK")
			return
		}

		// todo
		if operation == "bank_card_update" {

		}
		// todo
		if operation == "bank_card_del" {

		}

		//普通查询
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.BankPay, 0)
		db := mysql.DB
		if status, isE := c.GetPostForm("status"); isE == true {
			db = db.Where("status=?", status)
		}
		var total int
		//条件
		db.Model(model.BankPay{}).Count(&total)
		db = db.Model(&model.BankPay{}).Offset((page - 1) * limit).Limit(limit).Order("created desc")
		db.Find(&sl)
		ReturnDataLIst2000(c, sl, total)

	}

	if action == "add" {
		name := c.PostForm("name")
		countryId, _ := strconv.Atoi(c.PostForm("country_id"))
		err := mysql.DB.Where("name=?  and  country_id=?", name, countryId).First(&model.BankPay{}).Error
		if err == nil {
			client.ReturnErr101Code(c, "不要重复添加")
			return
		}
		err = mysql.DB.Save(&model.BankPay{CountryId: countryId, Name: name, Created: time.Now().Unix(), Status: 1}).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "OK")
		return
	}

	if action == "update" {
		id := c.PostForm("id")
		update := model.BankPay{}
		if status, isE := c.GetPostForm("status"); isE == true {
			update.Status, _ = strconv.Atoi(status)
		}
		if status, isE := c.GetPostForm("name"); isE == true {
			update.Name = status
		}
		if status, isE := c.GetPostForm("country_id"); isE == true {
			update.CountryId, _ = strconv.Atoi(status)
		}
		err := mysql.DB.Model(&model.BankPay{}).Where("id=?", id).Update(&update).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "OK")
		return
	}

}

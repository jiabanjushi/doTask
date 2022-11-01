package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"strconv"
	"strings"
	"time"
)

// OperationPayChannels 支付
func OperationPayChannels(c *gin.Context) {
	action := c.Query("action")

	if action == "select" {
		//普通查询
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.PayChannels, 0)
		db := mysql.DB

		db = db.Where("kinds=?", 1)
		var total int
		//条件
		db.Model(model.PayChannels{}).Count(&total)
		db = db.Model(&model.PayChannels{}).Offset((page - 1) * limit).Limit(limit).Order("updated desc")
		db.Find(&sl)

		for i, channels := range sl {
			country := model.Country{}
			err := mysql.DB.Where("id=?", channels.CountryId).First(&country).Error
			if err == nil {
				sl[i].CountryName = country.CountryName
			}

			bp := model.BankPay{}
			err = mysql.DB.Where("id=?", channels.BankPayId).First(&bp).Error
			if err == nil {
				sl[i].BankPayIDName = bp.Name
			}
		}

		ReturnDataLIst2000(c, sl, total)
	}

	if action == "add" {
		pc := model.PayChannels{}
		pc.Created = time.Now().Unix()
		pc.Updated = time.Now().Unix()
		pc.Kinds = 1
		pc.Maintenance = 1
		pc.Status = 1
		pc.BackIp = strings.TrimSpace(c.PostForm("back_ip"))
		pc.PayInterval = strings.TrimSpace(c.PostForm("pay_interval"))
		pc.Key = strings.TrimSpace(c.PostForm("key"))
		pc.PayCode = strings.TrimSpace(c.PostForm("pay_code"))
		pc.Merchants = strings.TrimSpace(c.PostForm("merchants"))
		pc.BackUrl = viper.GetString("Config.ip") + "/pay/back/" + strings.TrimSpace(c.PostForm("back_url"))
		pc.OnLine, _ = strconv.Atoi(c.PostForm("on_line"))
		pc.CurrencySymbol = strings.TrimSpace(c.PostForm("currency_symbol"))
		pc.CountryId, _ = strconv.Atoi(c.PostForm("country_id"))
		pc.PayUrl = strings.TrimSpace(c.PostForm("pay_url"))
		pc.Name = c.PostForm("name")
		pc.PayType, _ = strconv.Atoi(c.PostForm("pay_type"))
		pc.ExchangeRate, _ = strconv.ParseFloat(c.PostForm("exchange_rate"), 64)
		pc.PrivateKey = c.PostForm("private_key")
		pc.PublicKey = c.PostForm("public_key")
		err := mysql.DB.Where("name=? and kinds=?", pc.Name, 1).First(&model.PayChannels{}).Error
		if err == nil {
			client.ReturnErr101Code(c, " 不要重复添加")
			return
		}
		err = mysql.DB.Save(&pc).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "OK")

		return
	}

	if action == "update" {
		id := c.PostForm("id")
		//状态单独修改
		if status, isE := c.GetPostForm("status"); isE == true {
			ST, _ := strconv.Atoi(status)
			err := mysql.DB.Model(&model.PayChannels{}).Where("id=?", id).Update(&model.PayChannels{Updated: time.Now().Unix(), Status: ST}).Error
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}

			client.ReturnSuccess2000Code(c, "OK")
			return
		}

		pc := model.PayChannels{}
		pc.Updated = time.Now().Unix()
		pc.Maintenance = 1
		pc.BackIp = strings.TrimSpace(c.PostForm("back_ip"))
		pc.PayInterval = strings.TrimSpace(c.PostForm("pay_interval"))
		pc.Key = strings.TrimSpace(c.PostForm("key"))
		pc.PayCode = strings.TrimSpace(c.PostForm("pay_code"))
		pc.Merchants = strings.TrimSpace(c.PostForm("merchants"))
		pc.BackUrl = viper.GetString("Config.ip") + "/pay/back/" + strings.TrimSpace(c.PostForm("back_url"))
		pc.OnLine, _ = strconv.Atoi(c.PostForm("on_line"))
		pc.CurrencySymbol = strings.TrimSpace(c.PostForm("currency_symbol"))
		pc.CountryId, _ = strconv.Atoi(c.PostForm("country_id"))
		pc.PayUrl = strings.TrimSpace(c.PostForm("pay_url"))
		pc.Name = c.PostForm("name")
		pc.PayType, _ = strconv.Atoi(c.PostForm("pay_type"))
		pc.Maintenance, _ = strconv.Atoi(c.PostForm("maintenance"))
		pc.ExchangeRate, _ = strconv.ParseFloat(c.PostForm("exchange_rate"), 64)
		err := mysql.DB.Model(&model.PayChannels{}).Where("id=?", id).Update(&pc).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "OK")
		return

	}

}

// OperationPaidChannels 代付
func OperationPaidChannels(c *gin.Context) {
	action := c.Query("action")
	if action == "select" {
		//普通查询
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.PayChannels, 0)
		db := mysql.DB

		db = db.Where("kinds=?", 2)
		var total int
		//条件
		db.Model(model.PayChannels{}).Count(&total)
		db = db.Model(&model.PayChannels{}).Offset((page - 1) * limit).Limit(limit).Order("updated desc")
		db.Find(&sl)
		for i, channels := range sl {
			country := model.Country{}
			err := mysql.DB.Where("id=?", channels.CountryId).First(&country).Error
			if err == nil {
				sl[i].CountryName = country.CountryName
			}

			bp := model.BankPay{}
			err = mysql.DB.Where("id=?", channels.BankPayId).First(&bp).Error
			if err == nil {
				sl[i].BankPayIDName = bp.Name
			}
		}

		ReturnDataLIst2000(c, sl, total)
	}

	if action == "add" {
		pc := model.PayChannels{}
		pc.Created = time.Now().Unix()
		pc.Updated = time.Now().Unix()
		pc.Kinds = 2
		pc.Maintenance = 1
		pc.Status = 1
		pc.BackIp = strings.TrimSpace(c.PostForm("back_ip"))
		pc.PayInterval = strings.TrimSpace(c.PostForm("pay_interval"))
		pc.Key = strings.TrimSpace(c.PostForm("key"))
		pc.PayCode = strings.TrimSpace(c.PostForm("pay_code"))
		pc.Merchants = strings.TrimSpace(c.PostForm("merchants"))
		pc.BackUrl = viper.GetString("Config.ip") + "/paid/back/" + strings.TrimSpace(c.PostForm("back_url"))
		pc.OnLine, _ = strconv.Atoi(c.PostForm("on_line"))
		pc.CurrencySymbol = strings.TrimSpace(c.PostForm("currency_symbol"))
		pc.CountryId, _ = strconv.Atoi(c.PostForm("country_id"))
		pc.PayUrl = strings.TrimSpace(c.PostForm("pay_url"))
		pc.Name = c.PostForm("name")
		pc.PayType, _ = strconv.Atoi(c.PostForm("pay_type"))
		pc.ExchangeRate, _ = strconv.ParseFloat(c.PostForm("exchange_rate"), 64)
		err := mysql.DB.Where("name=? and kinds=?", pc.Name, 1).First(&model.PayChannels{}).Error
		if err == nil {
			client.ReturnErr101Code(c, " 不要重复添加")
			return
		}
		err = mysql.DB.Save(&pc).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "OK")
		return
	}

	if action == "update" {
		id := c.PostForm("id")
		//状态单独修改
		if status, isE := c.GetPostForm("status"); isE == true {
			ST, _ := strconv.Atoi(status)
			err := mysql.DB.Model(&model.PayChannels{}).Where("id=?", id).Update(&model.PayChannels{Updated: time.Now().Unix(), Status: ST}).Error
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}

			client.ReturnSuccess2000Code(c, "OK")
			return
		}
		pc := model.PayChannels{}
		pc.Updated = time.Now().Unix()
		pc.Maintenance = 1
		pc.BackIp = strings.TrimSpace(c.PostForm("back_ip"))
		pc.PayInterval = strings.TrimSpace(c.PostForm("pay_interval"))
		pc.Key = strings.TrimSpace(c.PostForm("key"))
		pc.PayCode = strings.TrimSpace(c.PostForm("pay_code"))
		pc.Merchants = strings.TrimSpace(c.PostForm("merchants"))
		pc.BackUrl = viper.GetString("Config.ip") + "/paid/back/" + strings.TrimSpace(c.PostForm("back_url"))
		pc.OnLine, _ = strconv.Atoi(c.PostForm("on_line"))
		pc.CurrencySymbol = strings.TrimSpace(c.PostForm("currency_symbol"))
		pc.CountryId, _ = strconv.Atoi(c.PostForm("country_id"))
		pc.PayUrl = strings.TrimSpace(c.PostForm("pay_url"))
		pc.Name = c.PostForm("name")
		pc.PayType, _ = strconv.Atoi(c.PostForm("pay_type"))
		pc.BankPayId, _ = strconv.Atoi(c.PostForm("bank_pay_id"))
		pc.Maintenance, _ = strconv.Atoi(c.PostForm("maintenance"))
		pc.ExchangeRate, _ = strconv.ParseFloat(c.PostForm("exchange_rate"), 64)

		err := mysql.DB.Model(&model.PayChannels{}).Where("id=?", id).Update(&pc).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "OK")
		return

	}

}

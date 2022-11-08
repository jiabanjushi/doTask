package client

import (
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"github.com/wangyi/GinTemplate/tools"
	"time"
)

func SetBank(c *gin.Context) {
	who, _ := c.Get("who")
	whoMap := who.(model.User)
	//获取银行卡
	action := c.Query("action")
	if action == "get" {
		bp := make([]model.BankPay, 0)
		mysql.DB.Where("status=?", 1).Find(&bp)

		dataArray := make([]map[string]interface{}, 0)
		for _, pay := range bp {
			bb := make([]model.BankCard, 0)
			mysql.DB.Where("bank_pay_id=?", pay.ID).Find(&bb)
			for _, card := range bb {
				data := make(map[string]interface{})
				data["id"] = card.ID
				data["name"] = card.BankName
				data["code"] = card.BankCode
				dataArray = append(dataArray, data)
			}
		}

		ReturnSuccess2000DataCode(c, dataArray, "OK")
		return
	}

	if action == "set" {
		bc := model.BankCardInformation{}
		err1 := mysql.DB.Where("user_id=?", whoMap.ID).First(&bc).Error
		bankCard := c.PostForm("bank_card_id")
		phone := c.PostForm("phone")
		mail := c.PostForm("mail")
		card := c.PostForm("card")
		username := c.PostForm("username")

		//查询卡是否存在
		bb := model.BankCard{}
		err := mysql.DB.Where("id=?", bankCard).First(&bb).Error
		if err != nil {
			ReturnErr101Code(c, map[string]interface{}{"identification": "MysqlErr", "msg": MysqlErr})
			return
		}
		save := model.BankCardInformation{
			BankCode: bb.BankCode,
			BankName: bb.BankName,
			Updated:  time.Now().Unix(),
			Kinds:    1,
			Status:   1,
			//UserId:   whoMap.ID,
			Phone:    phone,
			Mail:     mail,
			Card:     card,
			Username: username,
		}

		if idCard, isE := c.GetPostForm("id_card"); isE == true {
			save.IdCard = idCard
		}

		if err1 != nil {
			//新增
			save.Created = time.Now().Unix()
			save.UserId = whoMap.ID
			err := mysql.DB.Save(&save).Error
			if err != nil {
				ReturnErr101Code(c, map[string]interface{}{"identification": "MysqlErr", "msg": MysqlErr})
				return
			}
		} else {
			//更新
			err1 := mysql.DB.Model(&model.BankCardInformation{}).Update(&save).Error
			if err1 != nil {
				ReturnErr101Code(c, map[string]interface{}{"identification": "MysqlErr", "msg": MysqlErr})
				return
			}

		}

		ReturnSuccess2000Code(c, "ok")
		return

	}

	if action == "getBank" {
		bc := model.BankCardInformation{}
		err := mysql.DB.Where("user_id=?", whoMap.ID).First(&bc).Error
		if err != nil {
			tools.JsonWrite(c, NoBank, nil, "ok")
			return
		}
		ReturnSuccess2000DataCode(c, bc, "ok")
		return

	}

}

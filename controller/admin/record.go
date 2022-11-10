package admin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mmdb"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"github.com/wangyi/GinTemplate/pay"
	"strconv"
	"strings"
	"time"
)

// OperationRecord 线上线下充值
func OperationRecord(c *gin.Context) {
	action := c.Query("action")
	who, _ := c.Get("who")
	whoMap := who.(model.Admin)
	ctName, _ := mmdb.GetCountryForIp(c.ClientIP())
	if action == "select" {
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.Record, 0)
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

		if order, isE := c.GetPostForm("username"); isE == true {
			usr := model.User{}
			err := mysql.DB.Where("username=?", order).First(&usr).Error
			if err == nil {
				db = db.Where("user_id=?", usr.ID)
			}

		}

		db = db.Where("kinds=? and  on_line=?", 2, c.PostForm("on_line"))

		if order, isE := c.GetPostForm("order_num"); isE == true {
			db = db.Where("order_num=?", order)
		}
		if order, isE := c.GetPostForm("three_order_num"); isE == true {
			db = db.Where("three_order_num=?", order)
		}
		if order, isE := c.GetPostForm("status"); isE == true {
			db = db.Where("status=?", order)
		}

		db.Model(model.Record{}).Count(&total)
		db = db.Model(&model.Record{}).Offset((page - 1) * limit).Limit(limit).Order("updated desc")
		db.Find(&sl)
		for i, record := range sl {
			cc := model.PayChannels{}
			mysql.DB.Where("id=?", record.PayChannelsId).First(&cc)
			sl[i].PayChannel = cc
			user := model.User{}
			mysql.DB.Where("id=?", record.UserId).First(&user)
			sl[i].UserName = user.Username
			sl[i].TopAgent = user.TopAgent
		}
		ReturnDataLIst2000(c, sl, total)
	}

	if action == "update" {
		//
		id := c.PostForm("id")
		status, _ := strconv.Atoi(c.PostForm("status"))
		re := model.Record{}
		err := mysql.DB.Where("id=?", id).First(&re).Error
		if err != nil {
			client.ReturnErr101Code(c, "修改的订单不存在")
			return
		}
		if status == re.Status {
			client.ReturnErr101Code(c, "不要重复修改")
			return
		}
		if status != 3 {
			client.ReturnErr101Code(c, "状态只能是拉回")
			return
		}

		pcc := model.PayChannels{}
		err = mysql.DB.Where("id=?", re.PayChannelsId).First(&pcc).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}

		am, _ := strconv.ParseFloat(c.PostForm("authenticity_money"), 64)
		//status ==3   已上分
		change := model.UserBalanceChange{UserId: re.UserId, ChangeMoney: re.Money, Kinds: 5, RecordId: re.ID, PayChannelsExchangeRate: pcc.ExchangeRate, AuthenticityMoney: am}
		_, err = change.UserBalanceChangeFunc(mysql.DB)
		if err != nil {
			client.ReturnErr101Code(c, err.Error())

			return
		}
		//日志Sprintf
		log := model.Log{Kinds: 3, Content: fmt.Sprintf("%s|拉回充值订单号:%s", whoMap.AdminUser, re.OrderNum), Country: ctName, Ip: c.ClientIP()}
		log.CreateLogger(mysql.DB)
		client.ReturnSuccess2000Code(c, "OK")
		return

	}

}

// OperationWithdraw 提现
func OperationWithdraw(c *gin.Context) {
	action := c.Query("action")
	who, _ := c.Get("who")
	whoMap := who.(model.Admin)
	ctName, _ := mmdb.GetCountryForIp(c.ClientIP())
	if action == "select" {

		operation := c.PostForm("operation")

		fmt.Println(operation == "getPaidChannels")
		if operation == "getPaidChannels" {
			BC := make([]model.BankCard, 0)
			mysql.DB.Where("bank_name=?", c.PostForm("bank_name")).Find(&BC)
			rD := make([]map[string]interface{}, 0)
			fmt.Println(BC)
			//rD = append(rD, map[string]interface{}{"id": cc.ID, "pay_type":1 , "name": ""})
			for _, i2 := range BC {
				cc := model.PayChannels{}
				err := mysql.DB.Where("bank_pay_id=?", i2.BankPayId).First(&cc).Error
				if err == nil {
					rD = append(rD, map[string]interface{}{"id": cc.ID, "pay_type": cc.PayType, "name": cc.Name})

				}
			}

			//查询是否有本地代付
			BC2 := make([]model.PayChannels, 0)
			mysql.DB.Where("pay_type=?", 1).First(&BC2)
			for _, card := range BC2 {
				rD = append(rD, map[string]interface{}{"id": card.ID, "pay_type": card.PayType, "name": card.Name})
			}

			client.ReturnSuccess2000DataCode(c, rD, "ok")
			return
		}

		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.Record, 0)
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
		db.Model(model.Record{}).Count(&total)
		db = db.Model(&model.Record{}).Offset((page - 1) * limit).Limit(limit).Order("updated desc")
		db.Find(&sl)
		for i, record := range sl {
			cc := model.PayChannels{}
			mysql.DB.Where("id=?", record.PayChannelsId).First(&cc)
			sl[i].PayChannel = cc
			user := model.User{}
			mysql.DB.Where("id=?", record.UserId).First(&user)
			sl[i].UserName = user.Username
			sl[i].TopAgent = user.TopAgent
			BankCardInformation := model.BankCardInformation{}
			mysql.DB.Where("user_id=?", record.UserId).First(&BankCardInformation)
			sl[i].BankCardInformation = BankCardInformation

		}
		ReturnDataLIst2000(c, sl, total)
	}

	if action == "update" {
		id := c.PostForm("id")
		status, _ := strconv.Atoi(c.PostForm("status"))
		re := model.Record{}
		err := mysql.DB.Where("id=?", id).First(&re).Error
		if err != nil {
			client.ReturnErr101Code(c, "修改的订单不存在")
			return
		}
		if re.Status == status {
			client.ReturnErr101Code(c, "不要重复修改")
			return
		}
		//驳回
		if status == 6 {
			if re.Status == 5 {
				client.ReturnErr101Code(c, "订单已经结算成功,不可以驳回")
				return
			}
			//驳回原因
			rejectReason := c.PostForm("reject_reason")
			change := model.UserBalanceChange{
				RejectReason: rejectReason,
				RecordId:     re.ID,
				ChangeMoney:  re.Money,
				Kinds:        6,
				UserId:       re.UserId,
			}
			_, err := change.UserBalanceChangeFunc(mysql.DB)
			if err != nil {

				client.ReturnErr101Code(c, err.Error())
				return
			}
			client.ReturnSuccess2000Code(c, "驳回成功")
			return

		}
		//审核成功
		if status == 2 {
			err := mysql.DB.Model(&model.Record{}).Where("id=?", id).Update(&model.Record{Status: status}).Error
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}
			//日志
			LO := model.Log{Country: ctName, Ip: c.ClientIP(), Content: fmt.Sprintf("%s|审核提现订单:%s,成功", whoMap.AdminUser, re.OrderNum), Kinds: 3}
			LO.CreateLogger(mysql.DB)
			client.ReturnSuccess2000Code(c, "审核成功")
			return
		}
		//拉回订单(代付订单)
		if status == 5 {
			//代付类型
			pT, _ := strconv.Atoi(c.PostForm("pay_type"))
			//本地代付
			if pT == 1 {
				//获取玩家的余额
				user := model.User{}
				err := mysql.DB.Where("id=?", re.UserId).First(&user).Error
				if err != nil {
					client.ReturnErr101Code(c, err.Error())
					return
				}
				change := model.UserBalanceChange{
					Kinds:       7,
					ChangeMoney: re.Money,
					RecordId:    re.ID,
				}
				_, err = change.UserBalanceChangeFunc(mysql.DB)
				if err != nil {
					client.ReturnErr101Code(c, err.Error())
					return
				}
				client.ReturnSuccess2000Code(c, "代付成功")
				return

			}
			db := mysql.DB
			id := c.PostForm("pay_channels_id")
			pc := model.PayChannels{}
			err := db.Where("id=?", id).First(&pc).Error
			if err != nil {
				client.ReturnErr101Code(c, "pay_channels_id 不存在")
				return
			}
			BankCardInformation := model.BankCardInformation{}
			mysql.DB.Where("user_id=?", re.UserId).First(&BankCardInformation)
			//BPay代付
			f := re.Money * (1 - re.ServiceCharge) / pc.ExchangeRate
			TransferAmount := strconv.FormatFloat(f, 'f', 2, 64)
			//传银行卡 和 用户名
			CN := BankCardInformation.Card
			Bc := BankCardInformation.BankCode
			if pT == 2 {
				ExtendedParams := "bankAccount^" + CN + "|bankCode^" + Bc
				//哥伦比亚
				if pc.ExtendedParams == "2" {
					username := BankCardInformation.Username
					phone := BankCardInformation.Phone
					IC := BankCardInformation.IdCard
					ExtendedParams = "|payeeName^" + username + "|payeePhone^" + phone + "|IDNo^" + IC
				}
				//	银行账号+银行编码+用户姓名+手机号码+身份证号码
				//创建代付订单
				paid := pay.BPaid{
					MerchantNo:      pc.Merchants,
					MerchantOrderNo: re.OrderNum,
					CountryCode:     pc.CountryCode,
					CurrencyCode:    pc.CurrencySymbol,
					TransferType:    pc.PayCode,
					FeeDeduction:    "1",
					Remark:          "remark",
					ExtendedParams:  ExtendedParams,
					PayUrl:          pc.PayUrl,
					PrivateKey:      pc.PrivateKey,
					PublicKey:       pc.PublicKey, TransferAmount: TransferAmount, NotifyUrl: pc.BackUrl}
				_, err = paid.CreatedPaidOrder(mysql.DB)
				if err != nil {
					client.ReturnErr101Code(c, err.Error())
					return
				}
				//修改订单的状态为代付中
				mysql.DB.Model(&model.Record{}).Where("id=?", re.ID).Update(&model.Record{Status: 3, Updated: time.Now().Unix(), PayChannelsId: pc.ID})
				client.ReturnSuccess2000Code(c, "ok")
				return

			}

			//LrPay
			if pT == 3 {
				config := model.Config{}
				mysql.DB.Where("id=?", 1).First(&config)
				paid := pay.LrPid{
					Summary:        "remark",
					BankCode:       Bc,
					AccName:        BankCardInformation.Username,
					MerNo:          pc.Merchants,
					Province:       BankCardInformation.IdCard,
					ExtendedParams: pc.ExtendedParams,
					OrderAmount:    TransferAmount,
					MobileNo:       BankCardInformation.Phone,
					AccNo:          BankCardInformation.Card,
					NotifyUrl:      pc.BackUrl,
					CcyNo:          pc.CurrencySymbol,
					MerOrderNo:     re.OrderNum,
					PrivateKey:     pc.PrivateKey,
					PhpUrl:         config.PhpUrl,
					PayUrl:         pc.PayUrl,
				}
				_, err := paid.CreatedOrderLrPaid()
				if err != nil {
					client.ReturnErr101Code(c, err.Error())
					return
				}
				mysql.DB.Model(&model.Record{}).Where("id=?", re.ID).Update(&model.Record{Status: 3, Updated: time.Now().Unix(), PayChannelsId: pc.ID})
				client.ReturnSuccess2000Code(c, "ok")
				return

			}

			client.ReturnErr101Code(c, "选择代付")
			return

		}

	}

}

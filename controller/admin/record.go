package admin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mmdb"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"strconv"
	"strings"
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

		db = db.Where("kinds=? and  on_line=?", 2, c.PostForm("on_line"))
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

		}

	}

}

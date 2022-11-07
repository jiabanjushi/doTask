package client

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"github.com/wangyi/GinTemplate/tools"
	"strconv"
	"strings"
)

// GetRecords 获取账单
func GetRecords(c *gin.Context) {
	who, _ := c.Get("who")
	whoMap := who.(model.User)
	kinds := c.PostForm("kinds")
	//1 提现  2 充值 3任务佣金
	limit, _ := strconv.Atoi(c.PostForm("limit"))
	page, _ := strconv.Atoi(c.PostForm("page"))
	records := make([]model.Record, 0)
	var total int
	db := mysql.DB
	db = db.Where("user_id=? and kinds=? ", whoMap.ID, kinds)
	db.Model(&model.Record{}).Count(&total)
	db = db.Offset((page - 1) * limit).Limit(limit).Order("updated desc")

	if created, isE := c.GetPostForm("start"); isE == true {
		end := c.PostForm("end")
		db = db.Where("updated >= ?  and  updated <=?", created, end)
	}

	db.Find(&records)
	data := make(map[string]interface{}, 0)
	data["data"] = records
	data["total"] = total
	ReturnSuccess2000DataCode(c, data, "ok")
	return
}

// Recharge  TODO 用户充值
func Recharge(c *gin.Context) {
	who, _ := c.Get("who")
	whoMap := who.(model.User)
	action := c.Query("action")
	if action == "getPay" {
		pc := make([]model.PayChannels, 0)
		mysql.DB.Where("kinds=? and status= ?", 1, 1).Find(&pc)
		ReturnSuccess2000DataCode(c, pc, "OK")
		return
	}

	//充值

	if action == "recharge" {
		money, mERR := strconv.ParseFloat(c.PostForm("money"), 64)
		if mERR != nil {
			ReturnErr101Code(c, map[string]interface{}{"identification": "MysqlErr", "msg": mERR.Error()})
			return
		}
		payChannelsId := c.PostForm("pay_channels_id")
		//判断渠道是否存在
		pc := model.PayChannels{}
		err := mysql.DB.Where("kinds=? and id=? and status=?", 1, payChannelsId, 1).First(&pc).Error
		if err != nil {
			ReturnErr101Code(c, map[string]interface{}{"identification": "MysqlErr", "msg": MysqlErr})
			return
		}

		//正在维护
		if pc.Maintenance == 2 {
			ReturnErr101Code(c, map[string]interface{}{"identification": "ChannelMaintenance", "msg": ChannelMaintenance})
			return
		}
		PayIntervalArray := strings.Split(pc.PayInterval, "-")
		minMoney := 100.00
		maxMoney := 99999.00
		if len(PayIntervalArray) == 2 {
			minMoney, _ = strconv.ParseFloat(PayIntervalArray[0], 64)
			maxMoney, _ = strconv.ParseFloat(PayIntervalArray[1], 64)
		}
		nowMoney := pc.ExchangeRate * money
		if nowMoney < minMoney || nowMoney > maxMoney {
			ReturnErr101Code(c, map[string]interface{}{"identification": "NotRechargeMoney", "msg": NotRechargeMoney})
			return
		}

		db := mysql.DB.Begin()
		//创建充值订单
		r := model.Record{UserId: whoMap.ID, Money: nowMoney, OnLine: pc.OnLine, PayChannelsId: pc.ID}
		_, err = r.CreatedRechargeOrder(mysql.DB)
		if err != nil {
			db.Rollback()
			ReturnErr101Code(c, map[string]interface{}{"identification": "MysqlErr", "msg": MysqlErr})
			return
		}

		//

		choose := model.PayChannelsChoose{Record: r, PayChannels: pc}
		pay, err := choose.ChoosePay(db)
		fmt.Println("错误信息")
		fmt.Println(err)
		if err != nil {
			db.Rollback()
			ReturnErr101Code(c, map[string]interface{}{"identification": "PayFail", "msg": PayFail})
			return
		}
		//发送充值请求
		db.Commit()
		data := make(map[string]interface{})
		data["url"] = pay
		ReturnSuccess2000DataCode(c, data, "ok")
		return

	}

}

// Withdraw 提现   TODO 用户提现
func Withdraw(c *gin.Context) {

	action := c.Query("action")
	if action == "getHand" {
		config := model.Config{}
		mysql.DB.Where("id=?", 1).First(&config)
		ReturnSuccess2000DataCode(c, config.WithdrawalHand, "OK")
		return
	}

	// 用户提现
	who, _ := c.Get("who")
	whoMap := who.(model.User)
	money, _ := strconv.ParseFloat(c.PostForm("money"), 64)

	//判断用户是否所还有任务在做
	GetTask := model.GetTask{}
	err3 := mysql.DB.Where("user_id=?", whoMap.ID).Last(&GetTask).Error
	if err3 == nil {
		if GetTask.Status != 2 {
			ReturnErr101Code(c, map[string]interface{}{"identification": "CantWithdraw", "msg": CantWithdraw})
			return
		}
	}

	//判断是否已经绑定了银行卡
	err2 := mysql.DB.Where("user_id=?", whoMap.ID).First(&model.BankCardInformation{}).Error
	if err2 != nil {
		ReturnErr101Code(c, map[string]interface{}{"identification": "NoBindBankCard", "msg": NoBindBankCard})
		return
	}

	config := model.Config{}
	err := mysql.DB.Where("id=?", 1).First(&config).Error
	if err != nil {
		ReturnErr101Code(c, map[string]interface{}{"identification": "MysqlErr", "msg": err.Error()})
		return
	}

	if money < config.SystemMinWithdrawal {
		tools.JsonWrite(c, SystemMinWithdrawal, config.SystemMinWithdrawal, "err")
		return
	}

	if money > whoMap.Balance {
		ReturnErr101Code(c, map[string]interface{}{"identification": "NOEnoughMoney", "msg": NOEnoughMoney})
		return
	}

	//model.Record{UserId: whoMap.ID}
	change := model.UserBalanceChange{UserId: whoMap.ID, ChangeMoney: -money, Kinds: 4}
	changeFunc, _ := change.UserBalanceChangeFunc(mysql.DB)
	if changeFunc == -1 {
		ReturnErr101Code(c, map[string]interface{}{"identification": "NOEnoughMoney", "msg": NOEnoughMoney})
		return
	}
	if changeFunc == -2 {
		ReturnErr101Code(c, map[string]interface{}{"identification": "MysqlErr", "msg": MysqlErr})
		return
	}

	ReturnSuccess2000Code(c, "OK")
	return
}

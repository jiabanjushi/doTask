package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"strconv"
	"time"
)

// OperationFirstPage 数据首页操作
func OperationFirstPage(c *gin.Context) {
	action := c.Query("action")
	//who, _ := c.Get("who")
	//whoMap := who.(model.Admin)
	//ctName, _ := mmdb.GetCountryForIp(c.ClientIP())
	if action == "select" {
		operation := c.PostForm("operation")
		//首页数据
		if operation == "firstPage" {
			type Sta struct {
				RegisterNum       int     `json:"register_num"`
				LoginNum          int     `json:"login_num"`
				WithdrawMoney     float64 `json:"withdraw_money"`      //提现金额
				WithdrawNum       int     `json:"withdraw_num"`        //提现人数
				RechargeBankMoney float64 `json:"recharge_bank_money"` //银行充值金额
				RechargeBankNum   int     `json:"recharge_bank_num"`   //银行充值人数
				RechargeUsdtMoney float64 `json:"recharge_usdt_money"` //USDT充值金额
				RechargeUsdtNum   int     `json:"recharge_usdt_num"`   //USDT充值人数
				FirstRechargeNum  int     `json:"first_recharge_num"`  //首冲人数

			}
			type Data struct {
				Today model.Statistics
				All   Sta
			}
			var data Data
			today := model.Statistics{}
			mysql.DB.Where("data=?", time.Now().Format("2006-01-02")).First(&today)
			data.Today = today
			//总统计
			//RegisterNum
			mysql.DB.Model(&model.Statistics{}).Raw("SELECT SUM(register_num) as register_num  FROM statistics").Scan(&data.All)
			mysql.DB.Model(&model.Statistics{}).Raw("SELECT SUM(login_num) as login_num  FROM statistics").Scan(&data.All)
			mysql.DB.Model(&model.Statistics{}).Raw("SELECT SUM(withdraw_money) as withdraw_money  FROM statistics").Scan(&data.All)
			mysql.DB.Model(&model.Statistics{}).Raw("SELECT SUM(recharge_bank_money) as recharge_bank_money  FROM statistics").Scan(&data.All)
			mysql.DB.Model(&model.Statistics{}).Raw("SELECT SUM(recharge_bank_num) as recharge_bank_num  FROM statistics").Scan(&data.All)
			mysql.DB.Model(&model.Statistics{}).Raw("SELECT SUM(recharge_usdt_money) as recharge_usdt_money  FROM statistics").Scan(&data.All)
			mysql.DB.Model(&model.Statistics{}).Raw("SELECT SUM(recharge_usdt_num) as recharge_usdt_num  FROM statistics").Scan(&data.All)
			mysql.DB.Model(&model.Statistics{}).Raw("SELECT SUM(recharge_usdt_num) as recharge_usdt_num  FROM statistics").Scan(&data.All)
			mysql.DB.Model(&model.Statistics{}).Raw("SELECT SUM(first_recharge_num) as first_recharge_num  FROM statistics").Scan(&data.All)
			client.ReturnSuccess2000DataCode(c, data, "ok")
			return

		}
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.Statistics, 0)
		db := mysql.DB
		var total int
		db.Model(model.Statistics{}).Count(&total)
		db = db.Model(&model.Statistics{}).Offset((page - 1) * limit).Limit(limit).Order("updated desc")
		db.Find(&sl)
		ReturnDataLIst2000(c, sl, total)

	}

	if action == "update" {
		sta := model.Statistics{Date: c.PostForm("date")}
		sta.UpdatedTodayData(mysql.DB)
		client.ReturnSuccess2000Code(c, "执行成功")
		return
	}

}

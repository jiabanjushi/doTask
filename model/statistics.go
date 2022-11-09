package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/wangyi/GinTemplate/tools"
	"go.uber.org/zap"
	"reflect"
	"strconv"
	"time"
)

// Statistics 每日统计
type Statistics struct {
	ID                int     `gorm:"primaryKey"`
	RegisterNum       int     `gorm:"comment:'注册人数'"`
	LoginNum          int     `gorm:"comment:'登录人数'"`
	WithdrawMoney     float64 //提现金额
	WithdrawNum       int     //提现人数
	RechargeBankMoney float64 //银行充值金额
	RechargeBankNum   int     //银行充值人数
	RechargeUsdtMoney float64 //USDT充值金额
	RechargeUsdtNum   int     //USDT充值人数
	FirstRechargeNum  int     //首冲人数
	Date              string
	Created           int64
	Updated           int64
}

// CheckIsExistModelStatistics 创建数据库
func CheckIsExistModelStatistics(db *gorm.DB) {
	if db.HasTable(&Statistics{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Statistics{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Statistics{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

// CreatedStatistics 创建或者更新每日数据
func (st *Statistics) CreatedStatistics(db *gorm.DB) {
	//这里要做并发限制(加读写锁)  --后期考虑
	//日期判断(今日是否已经存在了数据)
	sp := Statistics{}
	err := db.Where("date=? ", time.Now().Format("2006-01-02")).First(&sp).Error
	if err != nil {
		//今日数据不存在 创建
		st.Created = time.Now().Unix()
		st.Updated = time.Now().Unix()
		st.Date = time.Now().Format("2006-01-02")
		err := db.Save(&st).Error
		if err != nil {
			zap.L().Debug(err.Error())
		}
	} else {
		//今日数据存在
		st.Updated = time.Now().Unix()
		//注册人数数据更新
		var i interface{} = *st
		value := reflect.ValueOf(i)
		//注册人数数据更新
		RegisterNum := value.FieldByName("RegisterNum")
		if RegisterNum.Int() != 0 {
			err := db.Model(&Statistics{}).Where("id=?", sp.ID).Update("register_num", gorm.Expr("register_num + ?", 1)).Error
			if err != nil {
				zap.L().Debug(err.Error())
			}
		}
		//登录人数数据更新
		LoginNum := value.FieldByName("LoginNum")
		if LoginNum.Int() != 0 {
			err := db.Model(&Statistics{}).Where("id=?", sp.ID).Update("login_num", gorm.Expr("login_num + ?", 1)).Error
			if err != nil {
				zap.L().Debug(err.Error())
			}
		}

	}
}

// UpdatedTodayData 更新今日的数据
func (st *Statistics) UpdatedTodayData(db *gorm.DB) {

	type Name struct {
		Withdraw      float64 `json:"withdraw"`       //提现金额
		WithdrawNum   int     `json:"withdraw_num"`   //提现人数
		Recharge      float64 `json:"recharge"`       //充值金额
		RechargeNum   int     `json:"recharge_num"`   //充值个数
		FirstRecharge int     `json:"first_recharge"` //首冲
	}
	var name Name
	//提现金额 COUNT(*)
	db.Raw("select  SUM(money) as withdraw from records where status=? and kinds=? and date=?", 5, 1, st.Date).Scan(&name)
	//提现人数
	db.Raw("select   count(distinct user_id) as withdraw_num from records where status=? and kinds=? and date=? ", 5, 1, st.Date).Scan(&name)
	//银行充值金额
	db.Raw("select  SUM(money) as recharge from records where status=? and kinds=? and date=?", 3, 2, st.Date).Scan(&name)
	//银行充值人数
	db.Raw("select  count(distinct user_id) as recharge_num from records where status=? and kinds=? and date=?  ", 3, 2, st.Date).Scan(&name)

	//首冲任务数
	record := make([]Record, 0)
	db.Where("status=? and kinds=? and date=?", 3, 2, st.Date).Order("updated asc").Find(&record)
	var people []string
	for _, r := range record {
		//判断这个用户是否已经检查过了
		username := strconv.Itoa(r.UserId)
		if tools.IsArray(people, username) == true {
			continue
		}
		people = append(people, username)
		//判断这个用户今天之前是否充值?
		timeObj1, _ := time.Parse("2006-01-02", st.Date)
		//判断
		err := db.Where("status=? and kinds=?  and updated< ?", 3, 2, timeObj1.Unix()).First(&record).Error
		if err == nil {
			name.FirstRecharge++
		}
	}
	updateData := make(map[string]interface{})
	updateData["WithdrawMoney"] = name.Withdraw
	updateData["WithdrawNum"] = name.WithdrawNum
	updateData["RechargeBankMoney"] = name.Recharge
	updateData["RechargeBankNum"] = name.RechargeNum
	updateData["FirstRechargeNum"] = name.FirstRecharge
	updateData["Updated"] = time.Now().Unix()
	statistics := Statistics{}
	err := db.Model(&Record{}).Where("date=?", st.Date).First(&statistics).Error
	if err != nil {
		//新增
		db.Save(&Statistics{
			Date:              st.Date,
			WithdrawMoney:     name.Withdraw,
			WithdrawNum:       name.WithdrawNum,
			RechargeBankMoney: name.Recharge,
			RechargeBankNum:   name.RechargeNum,
			FirstRechargeNum:  name.FirstRecharge,
			Created:           time.Now().Unix(),
			Updated:           time.Now().Unix(),
		})
	} else {
		//修改
		db.Model(&Statistics{}).Where("id=?", statistics.ID).Update(updateData)
	}
}

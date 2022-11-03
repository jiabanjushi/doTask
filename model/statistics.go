package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"reflect"
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

}

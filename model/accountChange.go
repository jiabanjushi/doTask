package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

type AccountChange struct {
	ID            int `gorm:"primaryKey"`
	UserId        int
	Money         float64 `gorm:"type:decimal(10,2)"`
	OriginalMoney float64 `gorm:"type:decimal(10,2)"`
	NowMoney      float64 `gorm:"type:decimal(10,2)"`
	RecordId      int     `gorm:"default:0"` //佣金  充值  提现
	TaskOrderId   int     `gorm:"default:0"`
	Created       int64
	Kinds         int `gorm:"-"` // 1 提现  2 充值 3任务佣金  4任务冻结  5任务解冻  6提现失败(驳回)
}

func CheckIsExistModelAccountChange(db *gorm.DB) {
	if db.HasTable(&AccountChange{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&AccountChange{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&AccountChange{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

// CreatedAccountChange 创建账变
func (ac *AccountChange) CreatedAccountChange(db *gorm.DB) (bool, error) {
	ac.Created = time.Now().Unix()
	err := db.Save(ac).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

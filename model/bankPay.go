package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type BankPay struct {
	ID        int `gorm:"primaryKey"`
	Status    int `gorm:"default:1"` //状态 1  显示  2不显示
	CountryId int
	Name      string
	Created   int64
}

func CheckIsExistModelBankPay(db *gorm.DB) {
	if db.HasTable(&BankPay{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&BankPay{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&BankPay{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

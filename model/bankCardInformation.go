package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type BankCardInformation struct {
	ID       int `gorm:"primaryKey"`
	UserId   int
	Kinds    int    // 1银行卡 2U地址
	BankName string //银行卡名称
	BankCode string //银行卡编号
	Status   int    //状态 1正常  2禁用
	Card     string
	Username string
	Phone    string
	Mail     string
	Created  int64
	Updated  int64
}

func CheckIsExistModelBankCardInformation(db *gorm.DB) {
	if db.HasTable(&BankCardInformation{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&BankCardInformation{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&BankCardInformation{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

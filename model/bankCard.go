package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type BankCard struct {
	ID        int    `gorm:"primaryKey"`
	BankPayId int    //
	BankName  string //银行卡名称
	BankCode  string //银行卡编号
	Status    int    `gorm:"default:1"` //状态 1正常  2禁用
	Created   int64
	Updated   int64
}

func CheckIsExistModelBankCard(db *gorm.DB) {
	if db.HasTable(&BankCard{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&BankCard{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&BankCard{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

// CreatedBank 创建银行卡
func (B *BankCard) CreatedBank(db *gorm.DB) {

}

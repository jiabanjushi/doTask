package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

// PayChannels 支付/代付通道
type PayChannels struct {
	ID             int `gorm:"primaryKey"`
	Kinds          int
	PayUrl         string
	CountryId      int
	CurrencySymbol string
	OnLine         int
	BackUrl        string
	Merchants      string
	PayCode        string
	Key            string
	PayInterval    string
	BackIp         string
	BankPayId      int
	Created        int64
	Updated        int64
	Maintenance    int
	Status         int
	Name           string  //名字
	PayType        int     //1  USDT     2 印度lepay
	PublicKey      string  `gorm:"type:text"`
	PrivateKey     string  `gorm:"type:text"`
	ExchangeRate   float64 `gorm:"type:decimal(10,2)"`
	CountryName    string  `gorm:"-"`
	BankPayIDName  string  `gorm:"-"`
}

func CheckIsExistModelPayChannels(db *gorm.DB) {
	if db.HasTable(&PayChannels{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&PayChannels{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&PayChannels{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

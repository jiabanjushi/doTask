package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

type Country struct {
	ID          int `gorm:"primaryKey"`
	CountryName string
	Status      int `gorm:"状态1正常 2禁用"`
	Created     int64
	Updated     int64
}

// CheckIsExistModelCountry 创建数据库
func CheckIsExistModelCountry(db *gorm.DB) {
	if db.HasTable(&Country{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Country{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Country{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
		co := Country{CountryName: "Spain", Status: 1, Created: time.Now().Unix(), Updated: time.Now().Unix(), ID: 1}
		db.Save(&co)
	}
}

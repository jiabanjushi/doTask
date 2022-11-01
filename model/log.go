package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

type Log struct {
	ID      int    `gorm:"primaryKey"`
	Kinds   int    //日志类容 1 注册  2登录  3管理员操作日志
	Content string `gorm:"type:text"`
	Status  int    `gorm:"comment: '1正常的日志  2错误的日志 ';default:1"`
	Ip      string
	Country string
	Created int64
}

// CheckIsExistModelLog CheckIsExistModelUser 创建User
func CheckIsExistModelLog(db *gorm.DB) {
	if db.HasTable(&Log{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Log{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Log{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

// CreateLogger 创建日志
func (l *Log) CreateLogger(db *gorm.DB) {
	l.Created = time.Now().Unix()
	//日志入库
	db.Save(&l)
}

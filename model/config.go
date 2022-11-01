package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type Config struct {
	ID                  int     `gorm:"primaryKey"`
	InitializeBalance   float64 `gorm:"type:decimal(10,2);default:20;comment:'用户注册时候默认的余额'"`
	RequestLimit        int     `gorm:"default:1;comment:'用户针对同一个接口请求的频率/1秒'"`
	AdminGoogleStatus   int     `gorm:"default:1;comment:'管理员是否要登录是否要谷歌验证,status 1需要 2不需要'"`
	WithdrawalHand      float64 `gorm:"type:decimal(10,2);default:0.00"`
	SystemMinWithdrawal float64 `gorm:"type:decimal(10,2);default:100.00"`
	AutomaticPoints     int
	TimeZone            string
}

// CheckIsExistModelConfig   创建Config
func CheckIsExistModelConfig(db *gorm.DB) {
	if db.HasTable(&Config{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Config{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Config{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
		db.Save(&Config{ID: 1})
	}
}

// GetInitializeBalance 获取用户初始化的余额
func GetInitializeBalance(db *gorm.DB) float64 {
	c := Config{}
	err := db.Where("id=?", 1).First(&c).Error
	if err != nil {
		return 20
	}
	return c.InitializeBalance
}

func GetConfigAdminGoogleStatus(db *gorm.DB) int {
	c := Config{}
	err := db.Where("id=?", 1).First(&c).Error
	if err != nil {
		return 2
	}
	return c.AdminGoogleStatus
}

package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type Slideshow struct {
	ID        int `gorm:"primaryKey"`
	ImageUrl  string
	Status    int `gorm:"状态1正常 2禁用"`
	CountryId int `gorm:"国家id"`
	Remark    string
	Created   int64
	Updated   int64
	Country   string `gorm:"-"`
}

// CheckIsExistModelSlideshow 创建数据库
func CheckIsExistModelSlideshow(db *gorm.DB) {
	if db.HasTable(&Slideshow{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Slideshow{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Slideshow{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

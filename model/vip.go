package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

type Vip struct {
	ID      int    `gorm:"primaryKey"`
	Name    string //会员名字
	Created int64  //创建时间
}

// CheckIsExistModelVip CheckIsExistModelConfig   创建Config
func CheckIsExistModelVip(db *gorm.DB) {
	if db.HasTable(&Vip{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Vip{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Vip{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}

		db.Save(&Vip{ID: 1, Name: "SVIP0", Created: time.Now().Unix()})
		db.Save(&Vip{ID: 2, Name: "SVIP1", Created: time.Now().Unix()})
		db.Save(&Vip{ID: 3, Name: "SVIP2", Created: time.Now().Unix()})
		db.Save(&Vip{ID: 4, Name: "SVIP3", Created: time.Now().Unix()})
		db.Save(&Vip{ID: 5, Name: "SVIP4", Created: time.Now().Unix()})
		db.Save(&Vip{ID: 6, Name: "SVIP5", Created: time.Now().Unix()})
		db.Save(&Vip{ID: 8, Name: "SVIP6", Created: time.Now().Unix()})
		db.Save(&Vip{ID: 8, Name: "SVIP7", Created: time.Now().Unix()})
		db.Save(&Vip{ID: 9, Name: "SVIP8", Created: time.Now().Unix()})
		db.Save(&Vip{ID: 10, Name: "SVIP9", Created: time.Now().Unix()})

	}
}

func (v *Vip) IsExist(db *gorm.DB) bool {
	err := db.Where("id=?", v.ID).First(&Vip{}).Error
	if err == nil {
		return true
	}
	return false
}

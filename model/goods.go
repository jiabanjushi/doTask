package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type Goods struct {
	ID          int    `gorm:"primaryKey"`
	GoodsName   string `gorm:"comment:'商品名称'"` //日志类容 1 注册  2登录
	GoodsImages string `gorm:"comment:'商品图片'"`
	Count       int    `gorm:"default:0"` //   获取的次数
	Created     int64
}

func CheckIsExistModelGoods(db *gorm.DB) {
	if db.HasTable(&Goods{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Goods{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Goods{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

// RandGetGoods 随机获取商品
func (g *Goods) RandGetGoods(db *gorm.DB) (int, error) {
	err := db.Order("count asc").Limit(1).First(g).Error
	if err != nil {
		return 0, err
	}
	db.Model(&Goods{}).Where("id=?", g.ID).Update(&Goods{Count: g.Count + 1})
	return g.ID, nil
}

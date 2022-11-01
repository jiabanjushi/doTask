package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	eeor "github.com/wangyi/GinTemplate/error"
	"github.com/wangyi/GinTemplate/tools"
	"time"
)

type Admin struct {
	ID             int    `gorm:"primaryKey"`
	AdminUser      string `gorm:"comment:'管理员账号'"`
	Password       string
	Nickname       string `gorm:"comment:'管理员昵称'"`
	Status         int    `gorm:"comment:'状态 1正常 2禁用'"`
	RoleId         int
	WhiteIps       string
	GoogleCode     string
	Token          string //38位
	Created        int64
	Updated        int64
	RoleName       string `gorm:"-"`
	AgencyUsername string //代理用户名
}

func CheckIsExistModelAdmin(db *gorm.DB) {
	if db.HasTable(&Admin{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Admin{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Admin{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
		//初始化数据
		ad := Admin{AdminUser: "admin", Password: "admin", Nickname: "SuperAdmin", RoleId: 1}
		ad.Create(db)
	}
}

func SetToken(db *gorm.DB) string {
	for i := 0; i < 5; i++ {
		token := tools.RandStringRunes(38)
		err := db.Where("token=?", token).First(&Admin{}).Error
		if err != nil {
			return token
		}
	}
	return ""
}

func (a *Admin) Create(db *gorm.DB) (bool, error) {
	//判断这个用户是否存在
	err := db.Where("admin_user=?", a.AdminUser).First(&Admin{}).Error
	if err == nil {
		return false, eeor.OtherError("该账户已经存在")
	}
	//生成token
	a.Token = SetToken(db)
	a.Status = 1
	a.Updated = time.Now().Unix()
	a.Created = time.Now().Unix()
	if a.Token == "" {
		return false, eeor.OtherError("系统繁忙,请稍后再试!")
	}
	err = db.Save(&a).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

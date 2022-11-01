package model

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

type Role struct {
	ID           int    `gorm:"primaryKey"`
	RoleName     string `gorm:"comment:'角色名称'"`
	Status       int    `gorm:"状态1正常 2禁用"`
	Jurisdiction string `gorm:"comment:'权限数据,json 字段进行存储' ;type:text"`
	Created      int64
	Updated      int64
}

type Secondary struct {
	Name       string //菜单的名字
	Status     int    //存在    状态 1 存在 2不存在
	RouterPath string
	Read       int //查询    状态 1 存在 2不存在
	Add        int //添加    状态 1 存在 2不存在
	Update     int //修改    状态 1 存在 2不存在
	Delete     int //删除    状态 1 存在 2不存在
}
type RoleMenus struct {
	Name   string      //等级菜单的名字
	Second []Secondary //耳机菜单
}

// CheckIsExistModelRole    创建Config
func CheckIsExistModelRole(db *gorm.DB) {
	if db.HasTable(&Role{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Role{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Role{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}

		//初始化超级管理员   遍历
		var ArrayRole []RoleMenus
		menus := make([]Menu, 0)
		db.Where("secondary=?", 0).Find(&menus) //获取所有的一级菜单
		for _, menu := range menus {
			//判断是否存在二级菜单
			var se []Secondary
			re := make([]Menu, 0)
			db.Where("secondary=?", menu.ID).Find(&re)
			for _, m := range re {
				se = append(se, Secondary{Name: m.Name, Status: 1, Read: 1, Add: 1, Update: 1, Delete: 1, RouterPath: m.RouterPath})
			}
			ro := RoleMenus{Name: menu.Name, Second: se}
			ArrayRole = append(ArrayRole, ro)
		}
		marshal, _ := json.Marshal(&ArrayRole)
		db.Save(&Role{ID: 1, RoleName: "超级管理员", Status: 1, Created: time.Now().Unix(), Updated: time.Now().Unix(), Jurisdiction: string(marshal)})
	}
}

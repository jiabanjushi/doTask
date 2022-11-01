package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

// Menu 菜单
type Menu struct {
	ID         int    `gorm:"primaryKey"`
	Name       string `gorm:"菜单名称"`
	Secondary  int    `gorm:"0 为顶级菜单  否则为2级菜单"`
	RouterPath string `gorm:"comment:'路由地址'"`
}

// CheckIsExistModelMenu 创建Menu
func CheckIsExistModelMenu(db *gorm.DB) {
	if db.HasTable(&Menu{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Menu{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Menu{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}

		//一级菜单
		db.Save(&Menu{ID: 1, Name: "首页", Secondary: 0, RouterPath: "firstPage"})
		db.Save(&Menu{ID: 2, Name: "会员管理", Secondary: 0, RouterPath: "user"})
		db.Save(&Menu{ID: 3, Name: "任务管理", Secondary: 0, RouterPath: "task"})
		db.Save(&Menu{ID: 4, Name: "财务管理", Secondary: 0, RouterPath: "money"})
		db.Save(&Menu{ID: 5, Name: "日志管理", Secondary: 0, RouterPath: "log"})
		db.Save(&Menu{ID: 6, Name: "数据分析", Secondary: 0, RouterPath: "data"})
		db.Save(&Menu{ID: 7, Name: "系统管理", Secondary: 0, RouterPath: "system"})
		//二级菜单(系统管理)
		db.Save(&Menu{ID: 711, Name: "用户角色", Secondary: 7, RouterPath: "system/admin"})
		db.Save(&Menu{ID: 712, Name: "角色管理", Secondary: 7, RouterPath: "system/role"})
		db.Save(&Menu{ID: 713, Name: "参数配置", Secondary: 7, RouterPath: "system/systemParameter"})
		db.Save(&Menu{ID: 714, Name: "国家管理", Secondary: 7, RouterPath: "system/country"})
		db.Save(&Menu{ID: 715, Name: "轮播图管理", Secondary: 7, RouterPath: "system/slideshow"})
		db.Save(&Menu{ID: 716, Name: "时区管理", Secondary: 7, RouterPath: "system/timezone"})
		//二级菜单(任务管理)
		db.Save(&Menu{ID: 311, Name: "任务分组", Secondary: 3, RouterPath: "task/group"})
		db.Save(&Menu{ID: 312, Name: "任务图片", Secondary: 3, RouterPath: "task/goods"})
		db.Save(&Menu{ID: 313, Name: "任务订单", Secondary: 3, RouterPath: "task/taskOrder"})
		//二级菜单(会员管理)
		db.Save(&Menu{ID: 211, Name: "普通会员", Secondary: 2, RouterPath: "user/user"})
		db.Save(&Menu{ID: 212, Name: "Vip列表", Secondary: 2, RouterPath: "user/vip"})
		db.Save(&Menu{ID: 213, Name: "顶级会员", Secondary: 2, RouterPath: "user/topUser"})
		//二级菜单(财务管理)
		db.Save(&Menu{ID: 411, Name: "支付管理", Secondary: 4, RouterPath: "money/pay"})
		db.Save(&Menu{ID: 412, Name: "代付管理", Secondary: 4, RouterPath: "money/anotherPay"})
		//db.Save(&Menu{ID: 413, Name: "绑卡审核", Secondary: 4, RouterPath: "money/bank"})
		db.Save(&Menu{ID: 414, Name: "线上充值", Secondary: 4, RouterPath: "money/onLineRecharge"})
		db.Save(&Menu{ID: 415, Name: "线下充值", Secondary: 4, RouterPath: "money/OfflineRecharge"})
		db.Save(&Menu{ID: 416, Name: "提现管理", Secondary: 4, RouterPath: "money/withdraw"})
		db.Save(&Menu{ID: 417, Name: "银行管理", Secondary: 4, RouterPath: "money/bank"})
		//日志
		db.Save(&Menu{ID: 511, Name: "登录日志", Secondary: 5, RouterPath: "log/login"})
		db.Save(&Menu{ID: 512, Name: "注册日志", Secondary: 5, RouterPath: "log/register"})
		db.Save(&Menu{ID: 513, Name: "管理操作", Secondary: 5, RouterPath: "log/adminOperation"})
		//数据分析
		db.Save(&Menu{ID: 611, Name: "数据首页", Secondary: 6, RouterPath: "data/firstPage"})
		db.Save(&Menu{ID: 612, Name: "每日数据", Secondary: 6, RouterPath: "data/everyday"})

	}
}

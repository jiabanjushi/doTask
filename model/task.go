package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	eeor "github.com/wangyi/GinTemplate/error"
	"time"
)

type Task struct {
	ID                int    `gorm:"primaryKey"`
	OverlayId         int    `gorm:"comment:'叠加id 默认为0  如果不为0 则为上级任务的id,这个是判断任务是否是叠加任务的根本'"`
	TaskName          string `gorm:"comment:'任务名称'"`
	TaskCount         int    //当任务为叠加的时候  则是显示有结果叠加任务
	AllCommissionRate string //佣金比例总统计 M@100;P@0.1;P@0.2 M代表固定金额   P代表百分号  分割符号 @  和 ;
	Dialog            int    `gorm:"comment:'是否有弹窗  1没有  2 要弹窗'"`
	DialogImage       string `gorm:"comment:'弹窗图片地址'"`
	OverlayIndex      int    `gorm:"comment:'叠加排序'"`
	VipId             int
	Created           int64
	Updated           int64
	PayMod            string `gorm:"comment:'支付模式'"`
	VipName           string `gorm:"-"`
	IfOverlay         bool   `gorm:"-"`
	TopAgent          string
}

func CheckIsExistModelTask(db *gorm.DB) {
	if db.HasTable(&Task{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Task{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Task{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

// CreateNoOverLayTask  创建非叠加任务
func (t *Task) CreateNoOverLayTask(db *gorm.DB) error {
	t.Created = time.Now().Unix()
	t.Updated = time.Now().Unix()
	//判断这个任务是否已经存在了
	err := db.Where("task_name=?", t.TaskName).First(&Task{}).Error
	if err == nil {
		return eeor.OtherError("不要重复添加")
	}
	if t.IfOverlay == true {
		// 查询已经存在的子任务

		db := db.Begin()
		var total int
		db.Model(&Task{}).Where("overlay_id=?", t.OverlayId).Count(&total)
		t.OverlayIndex = total + 1
		err = db.Save(t).Error
		if err != nil {
			return eeor.OtherError(err.Error())
		}
		//更新父级任务
		err := db.Model(&Task{}).Where("id=?", t.OverlayId).Update("task_count", gorm.Expr("task_count + ?", 1)).Error
		if err != nil {
			db.Rollback()
			return eeor.OtherError(err.Error())
		}
		db.Commit()
	} else {
		err = db.Save(t).Error
		if err != nil {
			return eeor.OtherError(err.Error())
		}
	}

	return nil
}

func (t *Task) GetList(db *gorm.DB, username string) []Task {
	tt := make([]Task, 0)
	if username != "" {
		db.Where("overlay_id=? and  vip_id=? and top_agent =?", 0, t.VipId, t.TopAgent).Find(&tt)
	} else {
		db.Where("overlay_id=? and  vip_id=? ", 0, t.VipId).Find(&tt)
	}
	return tt
}

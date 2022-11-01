package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	eeor "github.com/wangyi/GinTemplate/error"
	"strconv"
	"sync"
	"time"
)

type GetTask struct {
	ID          int `gorm:"primaryKey"`
	UserId      int
	TaskId      int
	Status      int     `gorm:"comment: '状态 1未完成 2已完成,3冻结';default:1"`
	Plan        float64 `gorm:"-"`
	TaskName    string  `gorm:"-"`
	Updated     int64
	Created     int64
	InitGetTask sync.Mutex //分配任务的时候上锁

}

func CheckIsExistGetTask(db *gorm.DB) {
	if db.HasTable(&GetTask{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&GetTask{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&GetTask{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

// GetTaskForUser 用户获取已经领取过的任务
func (gt *GetTask) GetTaskForUser(db *gorm.DB) []GetTask {
	grs := make([]GetTask, 0)
	db.Where("user_id=?  ", gt.UserId).Find(&grs)
	for i, gr := range grs {
		//查询id的名字
		ts := Task{}
		err := db.Where("id=?", gr.TaskId).First(&ts).Error
		if err == nil {
			grs[i].TaskName = ts.TaskName
		}
		//对于未完成的用户进行遍历  查看完成进度
		if gr.Status == 1 {
			//获取全部任务
			to := TaskOrder{GetTaskId: gr.ID}
			grs[i].Plan = to.GetPlan(db) //任务的完成进度
		}

	}
	return grs
}

// CreateGetTaskTable 创建 领取任务并且生产任务订单
func (gt *GetTask) CreateGetTaskTable(db *gorm.DB) (bool, error) {
	gt.InitGetTask.Lock()
	//重新查看,用户领取任务表
	defer gt.InitGetTask.Unlock()
	err := db.Where("(status=? or status=?) and user_id=?", 1, 3, gt.UserId).First(&GetTask{}).Error
	if err == nil {
		return false, eeor.OtherError("已经领取了任务,不要重复领取")
	}
	db = db.Begin()
	gt.Created = time.Now().Unix()
	gt.Updated = time.Now().Unix()
	gt.Status = 1
	err = db.Save(gt).Error
	if err != nil {
		db.Rollback()
		return false, err
	}

	//获取用户的余额
	user := User{}
	db.Where("id=?", gt.UserId).First(&user)
	order := TaskOrder{UserId: strconv.Itoa(gt.UserId), TaskId: gt.TaskId, GetTaskId: gt.ID, Money: user.Balance}
	_, err = order.InitTaskOrder(db)
	if err != nil {
		db.Rollback()
		return false, err
	}

	db.Commit()
	return true, nil
}

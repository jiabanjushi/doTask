package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/wangyi/GinTemplate/tools"
	"strconv"
	"strings"
	"time"
)

type TaskOrder struct {
	ID              int     `gorm:"primaryKey"`
	TaskId          int     `gorm:"comment:'任务id(任务分组id)'"`
	GetTaskId       int     //已领取的任务id
	GoodsId         int     //商品id
	TaskOrderNum    string  //任务单号开头 YJ
	OrderMoney      float64 `gorm:"type:decimal(10,2)"`
	CommissionRate  float64
	CommissionMoney float64 `gorm:"type:decimal(10,2)"`
	Status          int     // 状态 1已分配未领取 2已领取未完成 3已领取已完成 4提交任务中  5已提交,未结算
	PayPer          float64 //支付比例
	Created         int64
	Updated         int64
	GetAt           int64   //订单超时付款时间
	ClearAt         int64   //订单结算时间
	UserId          string  `gorm:"-"` //便于生成任务单号
	GoodsName       string  `gorm:"-"`
	GoodsImageUrl   string  `gorm:"-"`
	Dialog          int     `gorm:"-"`
	DialogImage     string  `gorm:"-"`
	Money           float64 `gorm:"-"`
	Uid             int     //玩家id
	Username        string  `gorm:"-"`
	TaskName        string  `gorm:"-"`
	TopAgent        string  `gorm:"-"` //顶级代理
}

func CheckIsExistModelTaskOrder(db *gorm.DB) {
	if db.HasTable(&TaskOrder{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&TaskOrder{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&TaskOrder{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

// GetPlan 通过  GetTaskId  获取 完成度()
func (t *TaskOrder) GetPlan(db *gorm.DB) float64 {
	//获取这个任务分组的任务总个数
	var allNUm int
	db.Model(&TaskOrder{}).Where("get_task_id=?", t.GetTaskId).Count(&allNUm)
	if allNUm == 0 {
		return 0
	}
	//获取已经完成的
	var one int
	db.Model(&TaskOrder{}).Where("get_task_id=? and status=?", t.GetTaskId, 3).Count(&one)
	if one == 0 {
		return 0
	}
	num1, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(one)/float64(allNUm)), 64) // 保留2位小数
	return num1
}

// InitTaskOrder 初始化 任务订单
func (t *TaskOrder) InitTaskOrder(db *gorm.DB) (bool, error) {
	//通过任务id 判断任务类型
	task := Task{}
	err := db.Where("id=?", t.TaskId).First(&task).Error
	if err != nil {
		return false, err
	}
	//叠加任务
	UID, _ := strconv.Atoi(t.UserId)

	fmt.Println()

	if task.OverlayIndex == 0 {
		ta := make([]Task, 0)
		db.Where("overlay_id=?", task.ID).Order("overlay_index asc").Find(&ta)
		for _, t2 := range ta {
			taskArray := strings.Split(t2.AllCommissionRate, ";")
			PayArray := strings.Split(t2.PayMod, ";")
			for i, s := range taskArray {
				ssArray := strings.Split(s, "@") //M 正常金额 P百分比
				gg := Goods{}
				goodsId, err := gg.RandGetGoods(db)
				if err != nil {
					return false, err
				}
				pp, _ := strconv.ParseFloat(PayArray[i], 64)
				newTask := TaskOrder{
					TaskId:       t2.ID,
					GetTaskId:    t.GetTaskId,
					GoodsId:      goodsId,
					TaskOrderNum: "YJ" + tools.RandStringRunesZM(2) + t.UserId + time.Now().Format("20060102150405") + strconv.Itoa(int(tools.GetRandomWithAll(100, 999))),
					OrderMoney:   0,
					Status:       1,
					Created:      time.Now().Unix(),
					Updated:      time.Now().Unix(),
					PayPer:       pp,
					Uid:          UID,
				}
				if ssArray[0] == "M" {
					float, _ := strconv.ParseFloat(ssArray[1], 64)
					newTask.CommissionMoney = float
					newTask.CommissionRate = 0
				}
				if ssArray[0] == "P" {
					float, _ := strconv.ParseFloat(ssArray[1], 64)
					newTask.CommissionRate = float
					newTask.CommissionMoney = 0
				}
				err = db.Save(&newTask).Error
				if err != nil {
					return false, err
				}
			}
		}

	} else {
		//普通任务
		taskArray := strings.Split(task.AllCommissionRate, ";")
		PayArray := strings.Split(task.PayMod, ";")
		//  获取余额
		for i, s := range taskArray {
			ssArray := strings.Split(s, "@") //M 正常金额 P百分比
			gg := Goods{}
			goodsId, err := gg.RandGetGoods(db)
			if err != nil {
				return false, err
			}
			pp, _ := strconv.ParseFloat(PayArray[i], 64)
			newTask := TaskOrder{
				TaskId:       t.TaskId,
				GetTaskId:    t.GetTaskId,
				GoodsId:      goodsId,
				TaskOrderNum: "YJ" + tools.RandStringRunesZM(2) + t.UserId + time.Now().Format("20060102150405") + strconv.Itoa(int(tools.GetRandomWithAll(100, 999))),
				OrderMoney:   0,
				Status:       1,
				Created:      time.Now().Unix(),
				Updated:      time.Now().Unix(),
				PayPer:       pp,
				Uid:          UID,
			}
			if ssArray[0] == "M" {
				float, _ := strconv.ParseFloat(ssArray[1], 64)
				newTask.CommissionMoney = float
				newTask.CommissionRate = 0
			}
			if ssArray[0] == "P" {
				float, _ := strconv.ParseFloat(ssArray[1], 64)
				newTask.CommissionRate = float
				newTask.CommissionMoney = 0
			}
			err = db.Save(&newTask).Error
			if err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

//CloseAnAccount 结算
func (t *TaskOrder) CloseAnAccount(db *gorm.DB) {
	//对 task_id 进行结算
	arrayTaskOrder := make([]TaskOrder, 0)
	db.Where("get_task_id=? and  task_id=?", t.GetTaskId, t.TaskId).Find(&arrayTaskOrder)
	for _, order := range arrayTaskOrder {
		userId, _ := strconv.Atoi(t.UserId)
		uc := UserBalanceChange{UserId: userId, TaskOrderId: order.ID, Kinds: 2}
		_, err := uc.UserBalanceChangeFunc(db)
		if err != nil {
			//订单结算失败  写进日志
			ll := Log{Kinds: 4, Created: time.Now().Unix(), Status: 2, Content: fmt.Sprintf("任务订单:%d,结算失败,失败原因:%s", t.ID, err.Error())}
			ll.CreateLogger(db)
			return
		}
		ll := Log{Kinds: 4, Created: time.Now().Unix(), Status: 2, Content: fmt.Sprintf("任务订单:%d,结算成功", t.ID)}
		ll.CreateLogger(db)
		//查看这个任务是否已经完成 ,如果已经完成就修改 get_task
	}
	var allCount int
	db.Model(&TaskOrder{}).Where("get_task_id=? ", t.GetTaskId).Count(&allCount)
	var alreadyOk int
	//获取已经完成的get_task_id
	db.Model(&TaskOrder{}).Where("get_task_id=?  and status=?", t.GetTaskId, 3).Count(&alreadyOk)
	if alreadyOk == allCount {
		db.Model(GetTask{}).Where("id=?", t.GetTaskId).Update(&GetTask{Updated: time.Now().Unix(), Status: 2})
	}

}

package client

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/dao/redis"
	"github.com/wangyi/GinTemplate/model"
	"github.com/wangyi/GinTemplate/tools"
	"strconv"
	"time"
)

// GetDoTask 获取可以做的任务
func GetDoTask(c *gin.Context) {
	who, _ := c.Get("who")
	whoMap := who.(model.User)
	db := mysql.DB
	//获取没有被领取的任务
	gt := model.GetTask{}
	err := db.Where("user_id=?", whoMap.ID).Last(&gt).Error
	if err != nil || gt.Status == 2 {
		ReturnErr101Code(c, map[string]interface{}{"identification": "UnassignedTask", "msg": UnassignedTask})
		return
	}

	err = db.Where("get_task_id=? and status= ? ", gt.ID, 6).First(&model.TaskOrder{}).Error
	if err == nil {
		ReturnErr101Code(c, map[string]interface{}{"identification": "TaskFrozen", "msg": TaskFrozen})
		return
	}

	//查看是否有正在结算的任务
	result, _ := redis.Rdb.HExists("Clearing_"+whoMap.Username, "start").Result()
	if result == true {
		tools.JsonWrite(c, TaskClearing, nil, "Ready to settle")
		return
	}

	//查看是否有已领取未完成的任务   有的话直接返回这个任务
	to := model.TaskOrder{}
	err = db.Where(" get_task_id=? and status=?", gt.ID, 2).First(&to).Error
	if err == nil {
		//返回这个任务
		gg := model.Goods{}
		err := db.Where("id=?", to.GoodsId).First(&gg).Error
		if err != nil {
			//对不起系统错误
			fmt.Println(1)
			ReturnErr101Code(c, map[string]interface{}{"identification": "GetTaskErr", "msg": GetTaskErr})
			return
		}
		to.GoodsImageUrl = gg.GoodsImages
		to.GoodsName = gg.GoodsName
		task := model.Task{}
		err = db.Where("id=?", to.TaskId).First(&task).Error
		if err != nil {
			//对不起系统错误
			fmt.Println(2)
			ReturnErr101Code(c, map[string]interface{}{"identification": "GetTaskErr", "msg": GetTaskErr})
			return
		}
		to.Dialog = task.Dialog
		to.DialogImage = task.DialogImage
		tools.JsonWrite(c, 2000, to, "ok")
		return
	}

	//这个用户没有领取过任务
	tp := model.TaskOrder{}
	err = db.Where("get_task_id=? and status=?", gt.ID, 1).Order("id asc").First(&tp).Error
	if err != nil {
		//对不起系统错误
		ReturnErr101Code(c, map[string]interface{}{"identification": "GetTaskErr", "msg": GetTaskErr})
		return
	}

	config := model.Config{}
	db.Where("id=?", 1).First(&config)

	//更新数据
	err = db.Model(&model.TaskOrder{}).Where("id=?", tp.ID).Update(&model.TaskOrder{
		Updated:    time.Now().Unix(),
		Status:     2,
		GetAt:      time.Now().Unix() + config.TaskTimeout,
		OrderMoney: model.GetRealBalance(db, gt.UserId) * tp.PayPer,
	}).Error

	if err != nil {
		ReturnErr101Code(c, map[string]interface{}{"identification": "GetTaskErr", "msg": GetTaskErr})
		return
	}
	newData := model.TaskOrder{}
	db.Where("id=?", tp.ID).First(&newData)
	gg := model.Goods{}
	err = db.Where("id=?", newData.GoodsId).First(&gg).Error
	if err != nil {
		//对不起系统错误
		ReturnErr101Code(c, map[string]interface{}{"identification": "GetTaskErr", "msg": GetTaskErr})
		return
	}
	newData.GoodsImageUrl = gg.GoodsImages
	newData.GoodsName = gg.GoodsName
	task := model.Task{}
	err = db.Where("id=?", newData.TaskId).First(&task).Error
	if err != nil {
		//对不起系统错误
		ReturnErr101Code(c, map[string]interface{}{"identification": "GetTaskErr", "msg": GetTaskErr})
		return
	}

	newData.Dialog = task.Dialog
	newData.DialogImage = task.DialogImage
	ReturnSuccess2000DataCode(c, newData, "ok")

	return

}

// SubmitTaskOrder 提交任务
func SubmitTaskOrder(c *gin.Context) {
	who, _ := c.Get("who")
	whoMap := who.(model.User)
	taskOderId := c.PostForm("task_order_id")
	payPassword := c.PostForm("pay_password")
	//判断任务状态是否OK
	db := mysql.DB
	//判断支付密码是否正确
	if whoMap.PayPassword != payPassword {
		ReturnErr101Code(c, map[string]interface{}{"identification": "ErrPayPassword", "msg": ErrPayPassword})
		return
	}
	taskOrder := model.TaskOrder{}
	err := db.Where("id=?", taskOderId).First(&taskOrder).Error
	if err != nil {
		ReturnErr101Code(c, map[string]interface{}{"identification": "OnFindTaskOrderId", "msg": OnFindTaskOrderId})
		return
	}
	if taskOrder.Status != 2 {
		ReturnErr101Code(c, map[string]interface{}{"identification": "DonDoubleCommit", "msg": DonDoubleCommit})
		return
	}
	//立刻修改改订单状态
	err = db.Model(&model.TaskOrder{}).Where("id=?", taskOrder.ID).Update(&model.TaskOrder{Updated: time.Now().Unix(), Status: 4}).Error
	if err != nil {
		tools.JsonWrite(c, MysqlErr, nil, err.Error())
		return
	}

	//UserBalanceChangeFunc
	change := model.UserBalanceChange{UserId: whoMap.ID, TaskOrderId: taskOrder.ID, ChangeMoney: -taskOrder.OrderMoney, Kinds: 1}
	changeFunc, err := change.UserBalanceChangeFunc(db)
	if changeFunc == -1 {
		db.Model(&model.TaskOrder{}).Where("id=?", taskOrder.ID).Update(&model.TaskOrder{Updated: time.Now().Unix(), Status: 2})
		//ReturnErr101Code(c, map[string]interface{}{"identification": "DotEnoughMoney", "msg": DotEnoughMoney})
		tools.JsonWrite(c, NoEnoughMoney, nil, err.Error())
		return
	}
	if changeFunc == -2 {
		tools.JsonWrite(c, MysqlErr, nil, err.Error())
		return
	}

	//这里判断我提交的这个任务订单是否 这组的最后一个?
	task := model.Task{}
	db.Where("id=?", taskOrder.TaskId).First(&task)
	//查询这个任务id  等待结算的个数
	var total int
	db.Model(&model.TaskOrder{}).Where("task_id=? and status=? and get_task_id=?", taskOrder.TaskId, 5, taskOrder.GetTaskId).Count(&total)

	if task.TaskCount == total {
		//进入结算
		taskOrderArray := make([]model.TaskOrder, 0)
		db.Where("task_id=? and get_task_id=?", taskOrder.TaskId, taskOrder.GetTaskId).Find(&taskOrderArray)

		config := model.Config{}
		db.Where("id=?", 1).First(&config)
		for i, order := range taskOrderArray {
			g := model.Goods{}
			db.Where("id=?", order.GoodsId).First(&g)
			taskOrderArray[i].GoodsName = g.GoodsName
			taskOrderArray[i].GoodsImageUrl = g.GoodsImages
			db.Model(model.TaskOrder{}).Where("id=?", order.ID).Update(&model.TaskOrder{ClearAt: time.Now().Unix() + config.SettlementWaitTime})
		}
		result := make(map[string]interface{})
		redis.Rdb.HSet("Clearing_"+whoMap.Username, "start", "doing")
		//进程
		go func() {
			time.Sleep(90 * time.Second)
			to := model.TaskOrder{UserId: strconv.Itoa(whoMap.ID), GetTaskId: taskOrder.GetTaskId, TaskId: taskOrder.TaskId}
			to.CloseAnAccount(mysql.DB)
			redis.Rdb.HDel("Clearing_"+whoMap.Username, "start")
		}()
		tools.JsonWrite(c, TaskClearing, result, "Ready to settle")
		return
	}
	//并且让用户继续 获取任务
	ReturnSuccess2000Code(c, "Submitted successfully")
	return
}

// GetTaskOrder 获取已经任务
func GetTaskOrder(c *gin.Context) {
	who, _ := c.Get("who")
	whoMap := who.(model.User)
	db := mysql.DB

	getOrder := make([]model.GetTask, 0)
	err := db.Where("user_id=?", whoMap.ID).Find(&getOrder).Error
	if err != nil {
		tools.JsonWrite(c, MysqlErr, nil, err.Error())
		return
	}
	status := c.PostForm("status")
	Order := make([]model.TaskOrder, 0)
	var ii []int
	for _, task := range getOrder {
		ii = append(ii, task.ID)
	}
	db.Where("status=? and get_task_id in  (?)", status, ii).Find(&Order)
	for i, order := range Order {
		g := model.Goods{}
		db.Where("id=?", order.GoodsId).First(&g)
		Order[i].GoodsName = g.GoodsName
		Order[i].GoodsImageUrl = g.GoodsImages

	}
	ReturnSuccess2000DataCode(c, Order, "ok")
	return

}

func GetTaskFirstPage(c *gin.Context) {
	type Data struct {
		TodayCommission            float64 `json:"today_commission"`              //今日佣金
		PersonalCommission         float64 `json:"personal_commission"`           //个人佣金
		AlreadyAccomplishTaskOrder int     `json:"already_accomplish_task_order"` //已经完成的订单
		FreezeNum                  int     `json:"freeze_num"`                    //冻结数量
		RemainingTaskOrderNum      int     `json:"remaining_task_order_num"`      //等待结算订单
		FreezeMoney                float64 `json:"freeze_money"`                  //冻结金额
		Balance                    float64 `json:"balance"`                       //当前余额
	}
	who, _ := c.Get("who")
	whoMap := who.(model.User)
	db := mysql.DB
	var dd Data
	//今日佣金
	db.Raw("select sum(money) as today_commission FROM records where kinds=? and user_id=? and date=?", 3, whoMap.ID, time.Now().Format("2006-01-02")).Scan(&dd)
	//个人佣金
	db.Raw("select sum(money) as personal_commission FROM records where kinds=? and user_id=? ", 3, whoMap.ID).Scan(&dd)
	//已经完成的订单
	db.Model(&model.TaskOrder{}).Where("status=? and  uid=?", 3, whoMap.ID).Count(&dd.AlreadyAccomplishTaskOrder)
	//冻结金额
	dd.FreezeMoney = whoMap.WorkingFreeze
	//当前余额
	dd.Balance = whoMap.Balance
	//等待结算订单
	db.Model(&model.TaskOrder{}).Where("status=? and  uid=?", 5, whoMap.ID).Count(&dd.RemainingTaskOrderNum)
	//冻结数量
	db.Model(&model.TaskOrder{}).Where("status=? and  uid=?", 6, whoMap.ID).Count(&dd.FreezeNum)
	ReturnSuccess2000DataCode(c, dd, "OK")
	return
}

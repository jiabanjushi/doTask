package process

import (
	"github.com/jinzhu/gorm"
	"github.com/robfig/cron"
	"github.com/wangyi/GinTemplate/model"
	"time"
)

// OrderTimeout 任务订单订单超时
func OrderTimeout(db *gorm.DB) {
	for true {
		order := make([]model.TaskOrder, 0)
		db.Where("status=? and get_at < ?", 2, time.Now().Unix()).Find(&order)
		for _, taskOrder := range order {
			db.Model(&model.TaskOrder{}).Where("id=?", taskOrder.ID).Update(&model.TaskOrder{Status: 6})
		}
		time.Sleep(5 * time.Second)
	}

}

// RechargeTimeout 充值订单超时
func RechargeTimeout(db *gorm.DB) {
	for true {
		order := make([]model.Record, 0)
		db.Where("status=? and kinds =?", 1, 2).Find(&order)
		for _, taskOrder := range order {
			if time.Now().Unix()-taskOrder.Created > 60*60 {
				db.Model(&model.Record{}).Where("id=?", taskOrder.ID).Update(&model.TaskOrder{Status: 4})
			}
		}
		time.Sleep(600 * time.Second)
	}
}

//  每段时间更新数据

func UpdateStatistics(db *gorm.DB) {
	for true {
		statistics := model.Statistics{Date: time.Now().Format("2006-01-02")}
		statistics.UpdatedTodayData(db)
		time.Sleep(2 * 60 * 60 * time.Second)
	}
}

// TimeTask 每日执行的任务(定时任务)
func TimeTask(db *gorm.DB) {
	c := cron.New()
	//{秒数} {分钟} {小时} {日期} {月份}
	c.AddFunc("0 0 1 * * ?", func() {
		statistics := model.Statistics{Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}
		statistics.UpdatedTodayData(db)
	})
	// 每个月一日早上六点运行
	c.Start()
	defer c.Stop()

}

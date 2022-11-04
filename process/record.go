package process

import (
	"github.com/jinzhu/gorm"
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
			time.Sleep(600 * time.Second)
		}
	}
}

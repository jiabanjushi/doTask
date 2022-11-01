package process

import (
	"github.com/jinzhu/gorm"
	"github.com/wangyi/GinTemplate/model"
	"time"
)

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

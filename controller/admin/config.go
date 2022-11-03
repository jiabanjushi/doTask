package admin

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/dao/redis"
	"github.com/wangyi/GinTemplate/model"
	"strconv"
	"time"
)

// OperationConfiguration  系统设置
func OperationConfiguration(c *gin.Context) {
	action := c.Query("action")
	if action == "select" {
		config := model.Config{}
		mysql.DB.Where("id=?", 1).First(&config)
		client.ReturnSuccess2000DataCode(c, config, "OK")
		return
	}

	if action == "update" {
		kinds, _ := strconv.Atoi(c.PostForm("kinds"))
		newConfig := model.Config{}
		mysql.DB.Where("id=?", 1).First(&newConfig)
		//基本设置
		if kinds == 1 {
			requestLimit, _ := strconv.Atoi(c.PostForm("request_limit"))
			adminGoogleStatus, _ := strconv.Atoi(c.PostForm("admin_google_status"))
			newConfig.RequestLimit = requestLimit
			newConfig.AdminGoogleStatus = adminGoogleStatus
			newConfig.SettlementWaitTime, _ = strconv.ParseInt(c.PostForm("settlement_wait_time"), 10, 64)
			newConfig.TaskTimeout, _ = strconv.ParseInt(c.PostForm("task_timeout"), 10, 64)
			newConfig.WebsiteH5 = c.PostForm("website_h5")
			//时区发生变化
			if newConfig.TimeZone != c.PostForm("time_zone") {
				loc, err := time.LoadLocation(c.PostForm("time_zone"))
				if err == nil {
					time.Local = loc // -> this is setting the global timezone
					fmt.Println(time.Now().Format("2006-01-02 15:04:05 "))
				}
			}
			newConfig.TimeZone = c.PostForm("time_zone")
		}

		//财务设置
		if kinds == 2 {
			newConfig.InitializeBalance, _ = strconv.ParseFloat(c.PostForm("initialize_balance"), 64)
			newConfig.WithdrawalHand, _ = strconv.ParseFloat(c.PostForm("withdrawal_hand"), 64)
			newConfig.SystemMinWithdrawal, _ = strconv.ParseFloat(c.PostForm("System_min_withdrawal"), 64)
			newConfig.AutomaticPoints, _ = strconv.Atoi(c.PostForm("automatic_points"))

		}
		err := mysql.DB.Model(&model.Config{}).Where("id=?", 1).Update(&newConfig).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		redis.Rdb.HMSet("SystemConfig", structs.Map(newConfig))
		client.ReturnSuccess2000Code(c, "ok")
		return

	}

}

package admin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mmdb"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"strconv"
	"strings"
)

// OperationUser 获取用户
func OperationUser(c *gin.Context) {
	action := c.Query("action")
	who, _ := c.Get("who")
	whoMap := who.(model.Admin)
	country, _ := mmdb.GetCountryForIp(c.ClientIP())

	if action == "select" {
		operation := c.PostForm("operation")
		//分配任务
		if operation == "get_task" {
			//用户的等级是否有资格被分配这个任务
			userId, _ := strconv.Atoi(c.PostForm("user_id"))
			user := model.User{}
			err := mysql.DB.Where("id=?", userId).First(&user).Error
			if err != nil {
				client.ReturnErr101Code(c, "用户不存在")
				return
			}

			taskId := c.PostForm("task_id")
			task := model.Task{}
			err = mysql.DB.Where("id=?", taskId).First(&task).Error
			if err != nil {
				client.ReturnErr101Code(c, "任务不存在")
				return
			}
			if task.VipId != user.VipId {
				client.ReturnErr101Code(c, "该会员不可以分配该任务,会员等级不匹配")
				return
			}

			//创建并且分配任务
			getTask := model.GetTask{UserId: userId, TaskId: task.ID}
			_, err = getTask.CreateGetTaskTable(mysql.DB)
			if err != nil {
				client.ReturnErr101Code(c, "分配任务失败:"+err.Error())
				return
			}
			LO := model.Log{Country: country, Ip: c.ClientIP(), Content: whoMap.AdminUser + "|" + "分配任务:" + taskId + "|用户:" + user.Username}
			LO.CreateLogger(mysql.DB)

			client.ReturnSuccess2000Code(c, "分配任务成功")
			return
		}
		//获取你可以领取的任务
		if operation == "task_list" {
			userId, _ := strconv.Atoi(c.PostForm("user_id"))
			user := model.User{}
			mysql.DB.Where("id= ?", userId).First(&user)
			task := model.Task{VipId: user.VipId, TopAgent: whoMap.AdminUser}
			taskArray := task.GetList(mysql.DB, whoMap.AgencyUsername)
			client.ReturnSuccess2000DataCode(c, taskArray, "ok")
			return
		}
		//获取历史任务
		if operation == "history" {

			user, _ := strconv.Atoi(c.PostForm("user_id"))
			task := model.GetTask{UserId: user}
			hs := task.GetTaskForUser(mysql.DB)
			client.ReturnSuccess2000DataCode(c, hs, "ok")
			return
		}

		//获取 用户的详情
		if operation == "detailOne" {
			userId := c.PostForm("user_id")
			ReData := make(map[string]interface{})
			user := model.User{}
			err := mysql.DB.Where("id=?", userId).First(&user).Error
			if err != nil {
				client.ReturnErr101Code(c, "用户不存在")
				return
			}
			bank := make([]model.BankCardInformation, 0)
			mysql.DB.Where("user_id=?", userId).Find(&bank)
			ReData["information"] = user
			ReData["bank"] = bank
			client.ReturnSuccess2000DataCode(c, ReData, "OK")
			return
		}

		//获取账变记录
		if operation == "changeMoney" {
			userId := c.PostForm("user_id")
			user := model.User{}
			err := mysql.DB.Where("id=?", userId).First(&user).Error
			if err != nil {
				client.ReturnErr101Code(c, "用户不存在")
				return
			}
			limit, _ := strconv.Atoi(c.PostForm("limit"))
			page, _ := strconv.Atoi(c.PostForm("page"))
			sl := make([]model.AccountChange, 0)
			db := mysql.DB
			db = db.Where("user_id=?", userId)
			var total int
			//条件
			db.Model(model.AccountChange{}).Count(&total)
			db = db.Model(&model.AccountChange{}).Offset((page - 1) * limit).Limit(limit).Order("created desc")
			db.Find(&sl)
			// 1 提现  2 充值 3任务佣金  4任务冻结  5任务解冻
			for i, change := range sl {
				if change.RecordId > 0 {
					record := model.Record{}
					mysql.DB.Where("id=?", change.RecordId).First(&record)
					sl[i].Kinds = record.Kinds
				}

				if change.TaskOrderId > 0 {
					TaskOrder := model.TaskOrder{}
					mysql.DB.Where("id=?", change.TaskOrderId).First(&TaskOrder)

					if change.NowMoney > change.OriginalMoney {
						sl[i].Kinds = 5
					} else {
						sl[i].Kinds = 4
					}

				}

			}

			ReturnDataLIst2000(c, sl, total)

			return
		}

		//加减余额
		if operation == "balance" {
			userId := c.PostForm("user_id")
			user := model.User{}
			err := mysql.DB.Where("id=?", userId).First(&user).Error
			if err != nil {
				client.ReturnErr101Code(c, "用户不存在")
				return
			}
			money := c.PostForm("money")
			float, _ := strconv.ParseFloat(money, 64)
			parseFloat, _ := strconv.Atoi(c.PostForm("kind"))
			change := model.UserBalanceChange{UserId: user.ID, ChangeMoney: float, Kinds: 3, RecordKind: parseFloat}
			_, err = change.UserBalanceChangeFunc(mysql.DB)
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}

			client.ReturnSuccess2000Code(c, "ok")
			return

		}

		//普通查询
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.User, 0)
		db := mysql.DB
		if whoMap.AgencyUsername != "" {
			arrayAU := strings.Split(whoMap.AgencyUsername, ",")
			db = db.Where("top_agent in  (?)", arrayAU)
		} else {
			db = db.Where("top_agent !=?", "")
		}

		if username, isExist := c.GetPostForm("username"); isExist == true {
			db = db.Where(" username like ?", "%"+username+"%")
		}
		var total int
		//条件
		db.Model(model.User{}).Count(&total)
		db = db.Model(&model.User{}).Offset((page - 1) * limit).Limit(limit).Order("created desc")
		db.Find(&sl)
		ReturnDataLIst2000(c, sl, total)

	}

	if action == "update" {
		userId := c.PostForm("user_id")
		user := model.User{}
		err := mysql.DB.Where("id=?", userId).First(&user).Error
		if err != nil {
			client.ReturnErr101Code(c, "用户不存在")
			return
		}

		update := make(map[string]interface{})

		//修改密码
		if status, isExist := c.GetPostForm("password"); isExist == true {
			update["Password"] = status
		}
		//修改支付密码
		if status, isExist := c.GetPostForm("pay_password"); isExist == true {
			update["PayPassword"] = status
		}

		//vip等级
		if status, isExist := c.GetPostForm("vip_id"); isExist == true {
			update["VipId"], _ = strconv.Atoi(status)
		}
		if status, isExist := c.GetPostForm("status"); isExist == true {
			update["Status"], _ = strconv.Atoi(status)
		}
		err = mysql.DB.Model(&model.User{}).Where("id=?", userId).Update(update).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "修改成功")
		return
	}

	if action == "add" {
		db := mysql.DB
		user := model.User{}
		user.Username = c.PostForm("username")
		err2 := mysql.DB.Where("username=?", user.Username).First(&model.User{}).Error
		if err2 == nil {
			client.ReturnErr101Code(c, "不可以重复添加")
			return
		}

		user.Password = c.PostForm("password")
		user.PayPassword = c.PostForm("pay_password")
		user.Kinds, _ = strconv.Atoi(c.PostForm("kinds"))
		if user.Kinds == 1 {
			user.InvitationCode, _ = client.CreateUserInvitationCode(db)
			if user.InvitationCode == "" {
				client.ReturnErr101Code(c, "邀请码生成失败")
				return
			}

		}
		user.Token, _ = client.CreateUserToken(db)
		if user.Token == "" {
			client.ReturnErr101Code(c, "token生成失败")
			return
		}
		user.VipId, _ = strconv.Atoi(c.PostForm("vip_id"))

		//判断添加的是否是 顶级代理  邀请码?
		if in, isX := c.GetPostForm("invitation_code"); isX == true {
			u := model.User{}
			err := mysql.DB.Where("invitation_code=?", in).First(&u).Error
			if err != nil {
				client.ReturnErr101Code(c, "无效邀请码")
				return
			}
			user.SuperiorAgent = u.Username

			if u.TopAgent == "" {
				user.TopAgent = u.Username
			} else {
				user.TopAgent = u.TopAgent

			}

			//等级树
			if u.LevelTree == "" {
				user.LevelTree = ";" + strconv.Itoa(u.ID) + ";"
			} else {
				user.LevelTree = u.LevelTree + strconv.Itoa(u.ID) + ";"
			}

		}

		err := db.Save(&user).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}

		client.ReturnSuccess2000Code(c, "添加数据成功")
		return

	}

}

// OperationTopUser 顶级会员获取
func OperationTopUser(c *gin.Context) {
	action := c.Query("action")
	//who, _ := c.Get("who")

	if action == "select" {
		//普通查询
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.User, 0)
		db := mysql.DB
		db = db.Where("top_agent=?", "")
		var total int
		//条件
		db.Model(&model.User{}).Count(&total)
		db = db.Model(&model.User{}).Offset((page - 1) * limit).Limit(limit).Order("created desc")
		db.Find(&sl)
		for i, user := range sl {
			var count int
			mysql.DB.Model(&model.User{}).Where("top_agent=?", user.Username).Count(&count)

			fmt.Println(user.Username)
			sl[i].NumberNum = count
		}

		ReturnDataLIst2000(c, sl, total)

	}
}

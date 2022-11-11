package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mmdb"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"github.com/wangyi/GinTemplate/tools"
	"strconv"
	"strings"
	"time"
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
			parseFloat, _ := strconv.Atoi(c.PostForm("kinds"))
			change := model.UserBalanceChange{UserId: user.ID, ChangeMoney: float, Kinds: 3, RecordKind: parseFloat}
			_, err = change.UserBalanceChangeFunc(mysql.DB)
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}

			client.ReturnSuccess2000Code(c, "ok")
			return

		}
		//获取邀请码  和链接
		if operation == "invite_code" {
			userArray := strings.Split(whoMap.AgencyUsername, ",")
			var data []map[string]string
			config := model.Config{}
			mysql.DB.Where("id=?", 1).First(&config)
			for _, i2 := range userArray {
				user := model.User{}
				err := mysql.DB.Where("username=?", i2).First(&user).Error
				if err == nil {
					data = append(data, map[string]string{"invite_code": user.InvitationCode, "url": config.WebsiteH5 + "?code=" + user.InvitationCode})
				}

			}
			client.ReturnSuccess2000DataCode(c, data, "ok")
			return
		}
		//获取银行列表
		if operation == "getBank" {
			bp := make([]model.BankPay, 0)
			mysql.DB.Where("status=?", 1).Find(&bp)
			dataArray := make([]map[string]interface{}, 0)
			var bankCarName []string
			for _, pay := range bp {
				bb := make([]model.BankCard, 0)
				mysql.DB.Where("bank_pay_id=?", pay.ID).Find(&bb)
				for _, card := range bb {
					//判断是否重复
					if tools.IsArray(bankCarName, card.BankName) == false {
						data := make(map[string]interface{})
						data["id"] = card.ID
						data["name"] = card.BankName
						data["code"] = card.BankCode
						dataArray = append(dataArray, data)
						bankCarName = append(bankCarName, card.BankName)
					}
				}
			}
			client.ReturnSuccess2000DataCode(c, dataArray, "OK")
			return
		}

		//添加银行卡
		if operation == "addBank" {
			userId := c.PostForm("user_id")
			//判断这个用户是否已经有银行卡了?
			err := mysql.DB.Where("user_id=?", userId).First(&model.BankCardInformation{}).Error
			if err == nil {
				client.ReturnErr101Code(c, "不要重复添加卡号")
				return
			}

			USERID, _ := strconv.Atoi(userId)
			addBank := model.BankCardInformation{
				UserId:   USERID,
				Kinds:    1,
				BankName: c.PostForm("bank_name"),
				BankCode: c.PostForm("bank_code"),
				Status:   1,
				Card:     c.PostForm("card"),
				Username: c.PostForm("username"),
				Phone:    c.PostForm("phone"),
				Mail:     c.PostForm("mail"),
				IdCard:   c.PostForm("id_card"),
				Created:  time.Now().Unix(),
				Updated:  time.Now().Unix(),
			}

			err = mysql.DB.Save(&addBank).Error
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}
			client.ReturnSuccess2000Code(c, "添加成功")
			return
		}

		//普通查询
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.User, 0)
		db := mysql.DB

		if topName, isE := c.GetPostForm("top_agent"); isE == true {
			db = db.Where("top_agent  =? ", topName)
		} else {
			if whoMap.AgencyUsername != "" {
				arrayAU := strings.Split(whoMap.AgencyUsername, ",")
				db = db.Where("top_agent in  (?)", arrayAU)
			} else {
				db = db.Where("top_agent !=?", "")
			}
		}

		if username, isExist := c.GetPostForm("username"); isExist == true {
			db = db.Where(" username like ?", "%"+username+"%")
		}
		var total int
		//条件
		db.Model(model.User{}).Count(&total)
		db = db.Model(&model.User{}).Offset((page - 1) * limit).Limit(limit).Order("created desc")
		db.Find(&sl)

		for i, user := range sl {
			gt := model.GetTask{}
			err := mysql.DB.Where("user_id=? and  status=?", user.ID, 1).First(&gt).Error
			if err == nil {
				task := model.Task{}
				mysql.DB.Where("id=?", gt.TaskId).First(&task)
				sl[i].DoingTask = task
			}

		}

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

		if status, isExist := c.GetPostForm("bank_card_information"); isExist == true {
			update := make(map[string]interface{})
			update["Phone"] = c.PostForm("phone")
			update["Mail"] = c.PostForm("mail")
			update["Kinds"], _ = strconv.Atoi(c.PostForm("Kinds"))
			update["BankCode"] = c.PostForm("bank_code")
			update["BankName"] = c.PostForm("bank_name")
			update["Card"] = c.PostForm("card")
			update["Username"] = c.PostForm("username")
			err := mysql.DB.Model(&model.BankCardInformation{}).Where("id=?", status).Update(update).Error
			if err != nil {
				client.ReturnErr101Code(c, "修改银行卡失败")
				return
			}
			client.ReturnSuccess2000Code(c, "修改成功")
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
		type DataR struct {
			AllMoney     float64 `json:"all_money"`      // 总余额
			AllNumberNum int     `json:"all_number_num"` //总的下级人数
			AllRecharge  float64 `json:"all_recharge"`   //所有的充值金额
			AllWithdraw  float64 `json:"all_withdraw"`   //所有的提现金额
			RechargeNum  int     `json:"recharge_num"`   //充值人数
			TaskDoNum    int     `json:"task_do_num"`    //做单人数
			WithdrawNum  int     `json:"withdraw_num"`   //提现人数
		}

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
			var data DataR
			//下级人数
			mysql.DB.Model(&model.User{}).Where("top_agent=?", user.Username).Count(&data.AllNumberNum)
			//总余额
			mysql.DB.Raw("SELECT SUM(balance) as  all_money FROM users where top_agent=? ", user.Username).Scan(&data)
			//所有的充值金额
			mysql.DB.Raw("SELECT SUM(records.money) as all_recharge FROM records  LEFT JOIN users  ON  users.id=records.user_id WHERE records.kinds=2 AND records.status=3 AND  users.top_agent=?", user.Username).Scan(&data)
			//所有的提现金额
			mysql.DB.Raw("SELECT SUM(records.money) as all_recharge FROM records  LEFT JOIN users  ON  users.id=records.user_id WHERE records.kinds=1 AND records.status=2 AND  users.top_agent=?", user.Username).Scan(&data)
			//recharge_num  充值的人数
			mysql.DB.Raw("SELECT count(*) as recharge_num FROM records  LEFT JOIN users  ON  users.id=records.user_id WHERE records.kinds=2 AND records.status=3 AND  users.top_agent=? GROUP BY records.user_id", user.Username).Scan(&data)
			//做单人数  task_do_num
			mysql.DB.Raw("SELECT count(*) as task_do_num  FROM task_orders  LEFT JOIN users  ON users.id=task_orders.uid WHERE users.top_agent=?  GROUP BY task_orders.uid", user.Username).Scan(&data)
			//提现人数
			mysql.DB.Raw("SELECT count(*)  as withdraw_num FROM records  LEFT JOIN users  ON  users.id=records.user_id WHERE records.kinds=1 AND records.status=2 AND  users.top_agent=?", user.Username).Scan(&data)
			sl[i].Extend = data
		}

		ReturnDataLIst2000(c, sl, total)

	}
}

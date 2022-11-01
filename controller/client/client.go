package client

import (
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/dao/mmdb"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/dao/redis"
	"github.com/wangyi/GinTemplate/model"
	"strconv"
	"time"
)

// Register 用户注册
func Register(c *gin.Context) {
	var re RegisterVerify
	//检查参数
	if err := c.ShouldBind(&re); err != nil {

		ReturnErr101Code(c, map[string]interface{}{"identification": "Verify", "msg": err.Error()})
		return
	}
	//先检查邀请码是否是存在的
	var u model.User
	err := mysql.DB.Where("invitation_code=?", re.InvitationCode).First(&u).Error
	if err != nil {
		//邀请码还不存在"Sorry, the invitation code does not exist
		ReturnErr101Code(c, map[string]interface{}{"identification": "invitation", "msg": RegisterErr01})
		return
	}
	//判断这个用户名是否已经存在了
	err = mysql.DB.Where("username=?", re.Username).First(&model.User{}).Error
	if err == nil {
		ReturnErr101Code(c, map[string]interface{}{"identification": "already", "msg": RegisterErr02})
		return
	}
	//生成用户的token
	q1, q2 := CreateUserToken(mysql.DB)
	if q2 == false {

		ReturnErr101Code(c, map[string]interface{}{"identification": "already", "msg": RegisterErr03})
		return
	}
	//生成用户的邀请码
	w1, w2 := CreateUserInvitationCode(mysql.DB)
	if w2 == false {
		ReturnErr101Code(c, map[string]interface{}{"identification": "already", "msg": RegisterErr03})
		return
	}
	var saveU model.User
	//获取用户的真实ip
	saveU.CreatedIp = c.ClientIP()
	saveU.Created = time.Now().Unix()
	country, err := mmdb.GetCountryForIp(saveU.CreatedIp)
	if err == nil {
		saveU.CreatedCountry = country
	}
	saveU.SuperiorAgent = u.Username
	if u.TopAgent == "" {
		saveU.TopAgent = u.Username
	} else {
		saveU.TopAgent = u.TopAgent
	}
	saveU.InvitationCode = w1
	saveU.Token = q1
	saveU.Username = re.Username
	saveU.Password = re.Password
	saveU.Phone = re.Phone
	saveU.PayPassword = re.PayPassword
	//等级树
	if u.LevelTree == "" {
		saveU.LevelTree = ";" + strconv.Itoa(u.ID) + ";"
	} else {
		saveU.LevelTree = u.LevelTree + strconv.Itoa(u.ID) + ";"
	}

	//获取系统的初始化金额
	saveU.Balance = model.GetInitializeBalance(mysql.DB)
	//准备注册入库
	err = mysql.DB.Save(&saveU).Error
	if err != nil {
		country, _ := mmdb.GetCountryForIp(c.ClientIP())
		Log := model.Log{Kinds: 1, Status: 2, Content: err.Error(), Ip: c.ClientIP(), Country: country}
		Log.CreateLogger(mysql.DB)
		return
	}
	//注册成功(写入日志)
	Log := model.Log{Kinds: 1, Status: 1, Content: saveU.Username + "|" + "注册成功", Ip: c.ClientIP(), Country: country}
	Log.CreateLogger(mysql.DB)
	//注册人数统计
	st := model.Statistics{RegisterNum: 1, TopAgent: saveU.TopAgent}
	st.CreatedStatistics(mysql.DB)
	ReturnSuccess2000Code(c, RegisterSuccess)
	return
}

// Login 用户登录
func Login(c *gin.Context) {
	var re LoginVerify
	//检查参数
	if err := c.ShouldBind(&re); err != nil {
		//ReturnVerifyErrCode(c, err)
		ReturnErr101Code(c, map[string]interface{}{"identification": "Verify", "msg": err.Error()})
		return
	}
	//检查账户密码是否正确
	user := model.User{}
	err := mysql.DB.Where("username=? and password =?", re.Username, re.Password).First(&user).Error
	if err != nil {
		ReturnErr101Code(c, map[string]interface{}{"identification": "loginPassword", "msg": LoginErr01})
		return
	}
	//修改登录的数据
	user.TheScLoginIp = user.TheLastLoginIp
	user.TheScLoginTime = user.TheLastLoginTime
	user.TheLastLoginIp = c.ClientIP()
	user.TheLastLoginTime = time.Now().Unix()
	//更新数据
	mysql.DB.Model(&model.User{}).Where("id=?", user.ID).Update(&user)
	//写入日志
	country, _ := mmdb.GetCountryForIp(c.ClientIP())
	lo := model.Log{Kinds: 2, Content: user.Username + "|" + "登录成功", Ip: c.ClientIP(), Country: country}
	lo.CreateLogger(mysql.DB)
	//更新日统计
	result, _ := redis.Rdb.HExists("LoginData", user.Username).Result()
	if result == false {
		st := model.Statistics{LoginNum: 1, TopAgent: user.TopAgent}
		st.CreatedStatistics(mysql.DB)
		redis.Rdb.HSet("LoginData", user.Username, 1)
	}
	//返回之前对一些数据做一些处理
	user.PayPassword = "******"
	user.Password = "******"
	user.VipName = "SVIP1" //默认值
	bip := model.Vip{}
	err = mysql.DB.Where("id=?", user.VipId).First(&bip).Error
	if err == nil {
		user.VipName = bip.Name
	}
	//登录成功
	ReturnSuccess2000DataCode(c, user, "ok")
	//设置token的有效期限
	redis.Rdb.Set("UserToken_"+user.Token, user.Username, 86400*time.Second)
	return

}

// GetInformation 获取个人信息
func GetInformation(c *gin.Context) {
	who, _ := c.Get("who")
	whoMap := who.(model.User)
	whoMap.PayPassword = "******"
	whoMap.Password = "********"
	vip := model.Vip{}
	mysql.DB.Where("id=?", whoMap.VipId).First(&vip)
	whoMap.VipName = vip.Name
	ReturnSuccess2000DataCode(c, whoMap, "ok")
	return

}

// GetMoneyInformation 获取   收益 冻结  余额   充值
func GetMoneyInformation(c *gin.Context) {

	type Data struct {
		Balance  float64 `json:"balance"`  //余额
		Income   float64 `json:"income"`   //总收益
		Recharge float64 `json:"recharge"` //充值
		Freeze   float64 `json:"freeze"`   //总收益
	}

	var data Data
	who, _ := c.Get("who")
	whoMap := who.(model.User)
	data.Balance = whoMap.Balance
	data.Freeze = whoMap.WorkingFreeze
	//总充值
	mysql.DB.Raw("SELECT SUM(money) as recharge FROM records where kinds=? and status=?", 2, 3).Scan(&data)
	mysql.DB.Raw("SELECT SUM(money) as income FROM records where kinds=? and status=?", 4, 1).Scan(&data)
	ReturnSuccess2000DataCode(c, data, "ok")
	return
}

//修改支付密码 或者 密码

func UpdatePassword(c *gin.Context) {
	who, _ := c.Get("who")
	whoMap := who.(model.User)
	if paw2, isE := c.GetPostForm("password"); isE == true {
		if paw1, isE := c.GetPostForm("old_password"); isE == true {
			if paw1 != whoMap.Password {
				ReturnErr101Code(c, map[string]interface{}{"identification": "PasswordErr", "msg": PasswordErr})
				return
			}
			err := mysql.DB.Model(&model.User{}).Where("id=?", whoMap.ID).Update(&model.User{Password: paw2}).Error
			if err != nil {

				ReturnErr101Code(c, map[string]interface{}{"identification": "MysqlErr", "msg": MysqlErr})
				return
			}
			ReturnSuccess2000Code(c, "ok")
			return

		}

	}

	if paw2, isE := c.GetPostForm("pay_password"); isE == true {
		if paw1, isE := c.GetPostForm("old_pay_password"); isE == true {
			if paw1 != whoMap.PayPassword {
				ReturnErr101Code(c, map[string]interface{}{"identification": "PasswordErr", "msg": PasswordErr})
				return
			}
			err := mysql.DB.Model(&model.User{}).Where("id=?", whoMap.ID).Update(&model.User{PayPassword: paw2}).Error
			if err != nil {
				ReturnErr101Code(c, map[string]interface{}{"identification": "MysqlErr", "msg": MysqlErr})
				return
			}
			ReturnSuccess2000Code(c, "ok")
			return

		}

	}

}

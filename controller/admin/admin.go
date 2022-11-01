package admin

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mmdb"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/dao/redis"
	"github.com/wangyi/GinTemplate/model"
	"github.com/wangyi/GinTemplate/tools"
	"strconv"
	"strings"
	"time"
)

// Login 管理员登录
func Login(c *gin.Context) {
	var lo LoginVerify
	//检查参数
	if err := c.ShouldBind(&lo); err != nil {
		client.ReturnVerifyErrCode(c, err)
		return
	}
	//首先还是判断账号密码是否错误
	ad := model.Admin{}
	err := mysql.DB.Where("admin_user=? and  password= ?", lo.Username, lo.Password).First(&ad).Error
	if err != nil {
		client.ReturnErr101Code(c, "登录失败,请检查密码或者账户或者谷歌验证码!")
		return
	}
	//判断系统设置是谷歌开关  1需要 2不需要
	gCSwitch := model.GetConfigAdminGoogleStatus(mysql.DB)
	if gCSwitch == 1 {
		if ad.GoogleCode == "" {
			//返回谷歌验证码让其绑定
			secret, _, qrCodeUrl := tools.InitAuth(lo.Username)
			err := mysql.DB.Model(&model.Admin{}).Where("id=?", ad.ID).Update(model.Admin{GoogleCode: secret}).Error
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}
			tools.JsonWrite(c, client.NeedGoogleBind, map[string]string{"codeUrl": qrCodeUrl}, "请先绑定谷歌账号")
			return
		}
		//校验谷歌验证
		verifyCode, _ := tools.NewGoogleAuth().VerifyCode(ad.GoogleCode, lo.GoogleCode)
		if !verifyCode {
			client.ReturnErr101Code(c, "谷歌验证失败")
			return
		}
	}
	//判断白名单
	if ad.WhiteIps != "" && strings.Index(ad.WhiteIps, c.ClientIP()) == -1 {
		client.ReturnErr101Code(c, "请添加白名单,ip:"+c.ClientIP())
		return
	}
	type ReturnData struct {
		Token string
		Menu  interface{}
	}
	var data ReturnData
	data.Token = ad.Token
	//查询角色
	role := model.Role{}
	err = mysql.DB.Where("id=?", ad.RoleId).First(&role).Error
	if err != nil {
		client.ReturnErr101Code(c, err.Error())
		return
	}
	var data2 interface{}
	err = json.Unmarshal([]byte(role.Jurisdiction), &data2)
	if err != nil {
		client.ReturnErr101Code(c, err.Error())
		return
	}
	data.Menu = data2
	client.ReturnSuccess2000DataCode(c, data, "ok")
	//写登录日志
	country, _ := mmdb.GetCountryForIp(c.ClientIP())
	L := model.Log{Content: ad.AdminUser + "|登录成功", Ip: c.ClientIP(), Country: country, Kinds: 2, Status: 1}
	L.CreateLogger(mysql.DB)
	redis.Rdb.Set("AdminToken_"+ad.Token, ad.AdminUser, 7*86400*time.Second)
	return
}

// OperationAdmin 管理员操作
func OperationAdmin(c *gin.Context) {
	action := c.Query("action")
	if action == "select" {
		//普通查询
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.Admin, 0)
		db := mysql.DB
		var total int
		//条件
		db.Model(model.Admin{}).Count(&total)
		db = db.Model(&model.Admin{}).Offset((page - 1) * limit).Limit(limit).Order("created desc")
		db.Find(&sl)

		for i, i2 := range sl {
			role := model.Role{}
			mysql.DB.Where("id=?", i2.RoleId).First(&role)
			sl[i].RoleName = role.RoleName

		}

		ReturnDataLIst2000(c, sl, total)

	}

	if action == "add" {
		admin := model.Admin{}
		admin.AdminUser = c.PostForm("admin_user")
		admin.Password = c.PostForm("password")
		admin.Nickname = c.PostForm("nickname")
		admin.RoleId, _ = strconv.Atoi(c.PostForm("role_id"))
		admin.WhiteIps = c.PostForm("white_ips")
		admin.GoogleCode = c.PostForm("google_code")
		admin.AgencyUsername = c.PostForm("agency_username")

		_, err := admin.Create(mysql.DB)
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "创建成功")
		return
	}

	if action == "update" {
		id := c.PostForm("id")
		admin := model.Admin{}
		admin.Updated = time.Now().Unix()
		if status, isE := c.GetPostForm("status"); isE == true {
			if id == "1" {
				client.ReturnErr101Code(c, "超管不能禁用")
				return
			}
			admin.Status, _ = strconv.Atoi(status)
			err := mysql.DB.Model(&model.Admin{}).Where("id=?", id).Update(&admin).Error
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}
			client.ReturnSuccess2000Code(c, "修改成功")
			return
		}

		admin.Password = c.PostForm("password")
		admin.Nickname = c.PostForm("nickname")
		admin.RoleId, _ = strconv.Atoi(c.PostForm("role_id"))
		admin.WhiteIps = c.PostForm("white_ips")
		admin.GoogleCode = c.PostForm("google_code")
		admin.AgencyUsername = c.PostForm("agency_username")
		err := mysql.DB.Model(&model.Admin{}).Where("id=?", id).Update(&admin).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "修改成功")
		return

	}

}

// OperationRole 角色操作
func OperationRole(c *gin.Context) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05 "))
	action := c.Query("action")
	if action == "select" {
		//普通查询
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.Role, 0)
		db := mysql.DB
		var total int
		//条件
		db.Model(model.Role{}).Count(&total)
		db = db.Model(&model.Role{}).Offset((page - 1) * limit).Limit(limit).Order("created desc")
		db.Find(&sl)

		ReturnDataLIst2000(c, sl, total)

	}

	if action == "update" {
		id := c.PostForm("id")
		admin := model.Role{}
		admin.Updated = time.Now().Unix()
		if status, isE := c.GetPostForm("status"); isE == true {

			admin.Status, _ = strconv.Atoi(status)
			err := mysql.DB.Model(&model.Role{}).Where("id=?", id).Update(&admin).Error
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}
			client.ReturnSuccess2000Code(c, "修改成功")
			return
		}

		admin.Jurisdiction = c.PostForm("jurisdiction")
		admin.RoleName = c.PostForm("role_name")
		err := mysql.DB.Model(&model.Role{}).Where("id=?", id).Update(&admin).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "修改成功")
		return

	}

	if action == "add" {
		admin := model.Role{}
		admin.Updated = time.Now().Unix()
		admin.Created = time.Now().Unix()
		admin.Jurisdiction = c.PostForm("jurisdiction")
		admin.RoleName = c.PostForm("role_name")
		if err := mysql.DB.Where("role_name=?", admin.RoleName).First(&model.Role{}).Error; err == nil {
			client.ReturnErr101Code(c, "不要重复添加角色")
			return
		}

		mysql.DB.Save(&admin)

		client.ReturnSuccess2000Code(c, "新增成功")
		return

	}
}

/**
 * @Author $
 * @Description
 * @Date $ $
 * @Param $
 * @return $
 **/
package router

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"github.com/wangyi/GinTemplate/controller/admin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/dao/redis"
	eeor "github.com/wangyi/GinTemplate/error"
	"github.com/wangyi/GinTemplate/model"
	"github.com/wangyi/GinTemplate/tools"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func Setup() *gin.Engine {

	//设置时区

	//生产模式
	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(Cors())
	r.Static("/static", "./static")
	r.Use(PermissionToCheck())
	r.Use(eeor.ErrHandler())
	r.NoMethod(eeor.HandleNotFound)
	r.Use(LimitIpRequestSameUrlForUser())
	//管理员接口
	administration := r.Group("/management/v1")
	{
		administration.POST("login", admin.Login)
		//系统管理
		{
			//轮播图
			administration.POST("system/slideshow", admin.OperationSlideshow)
			//国家管理
			administration.POST("system/country", admin.OperationCountry)
			//OperationConfiguration
			administration.POST("system/systemParameter", admin.OperationConfiguration)
			//OperationAdmin
			administration.POST("system/admin", admin.OperationAdmin)
			//OperationRole
			administration.POST("system/role", admin.OperationRole)

		}

		//任务管理
		{
			//任务图片管理   OperationGoods
			administration.POST("task/goods", admin.OperationGoods)
			//任务分组
			administration.POST("task/group", admin.OperationTask)
			//OperationTaskOder
			administration.POST("task/taskOrder", admin.OperationTaskOder)

		}

		//会员管理

		{
			//vip列表
			administration.POST("user/vip", admin.OperationVip)
			//获取用户 GetUsers
			administration.POST("user/user", admin.OperationUser)
			//topUser
			administration.POST("user/topUser", admin.OperationTopUser)

		}

		//财务管理
		{
			//OperationBank
			administration.POST("money/bank", admin.OperationBank)
			//OperationPayChannels
			administration.POST("money/pay", admin.OperationPayChannels)
			//OperationPaidChannels
			administration.POST("money/anotherPay", admin.OperationPaidChannels)
			//OperationRecord
			administration.POST("money/onLineRecharge", admin.OperationRecord)
			administration.POST("money/OfflineRecharge", admin.OperationRecord)
			//money/withdraw
			administration.POST("money/withdraw", admin.OperationWithdraw)

		}
		{
			//OperationLogger
			administration.POST("log/login", admin.OperationLogger)
			administration.POST("log/register", admin.OperationLogger)
			administration.POST("log/adminOperation", admin.OperationLogger)

		}

		{
			//OperationFirstPage
			administration.POST("data/firstPage", admin.OperationFirstPage)
			administration.POST("data/everyday", admin.OperationFirstPage)
		}

	}
	//用户接口
	user := r.Group("client/v1")
	{
		user.POST("register", client.Register)
		user.POST("login", client.Login)
		user.POST("getInformation", client.GetInformation)
		user.POST("getSlideshow", client.GetSlideshow)
		////GetMoneyInformation
		user.POST("getMoneyInformation", client.GetMoneyInformation)
		//UpdatePassword
		user.POST("updatePassword", client.UpdatePassword)

		//任务
		{
			//领取任务
			user.POST("task/getDoTask", client.GetDoTask)
			//提交任务 SubmitTaskOrder
			user.POST("task/submitTaskOrder", client.SubmitTaskOrder)
			//GetTaskOrder  查询任务
			user.POST("task/getTaskOrder", client.GetTaskOrder)
			//GetTaskFirstPage
			user.POST("task/getTaskFirstPage", client.GetTaskFirstPage)

		}

		//账单
		{
			//GetRecords
			user.POST("record/getRecords", client.GetRecords)
			//Recharge
			user.POST("record/recharge", client.Recharge)
			//Withdraw
			user.POST("record/withdraw", client.Withdraw)

		}

		{

			//SetBank
			user.POST("bank/setBank", client.SetBank)

		}

		user.POST("test", func(c *gin.Context) {

		})

	}
	//三方接口
	pay := r.Group("pay/v1")
	{
		pay.GET("")
	}

	r.Run(fmt.Sprintf(":%d", viper.GetInt("app.port")))
	return r
}

// PermissionToCheck 权限校验
func PermissionToCheck() gin.HandlerFunc {
	whiteUrl := []string{"/client/v1/register", "/client/v1/login", "/management/v1/login"}

	return func(c *gin.Context) {
		if !tools.IsArray(whiteUrl, c.Request.RequestURI) {
			//token  校验
			//判断是用户还是管理员
			fmt.Println(c.Request.URL.Path)
			token := c.Request.Header.Get("token")
			//用户
			if len(token) == 36 {
				ad := model.User{}
				err := mysql.DB.Where("token=?", token).First(&ad).Error
				if err != nil {
					tools.JsonWrite(c, client.IllegalityCode, nil, client.IllegalityMsg)
					c.Abort()
				}
				//判断token 是否过期?
				if redis.Rdb.Get("UserToken_"+token).Val() == "" {
					tools.JsonWrite(c, client.TokenExpire, nil, client.LoginExpire)
					c.Abort()
				}
				c.Set("who", ad)
				c.Next()
			} else if len(token) == 38 {
				//管理员
				ad := model.Admin{}
				err := mysql.DB.Where("token=?", token).First(&ad).Error
				if err != nil {
					tools.JsonWrite(c, client.IllegalityCode, nil, client.IllegalityMsg)
					c.Abort()
				}
				//判断token 是否过期?
				if redis.Rdb.Get("AdminToken_"+token).Val() == "" {
					tools.JsonWrite(c, client.TokenExpire, nil, client.LoginExpire)
					c.Abort()
				}
				//判断接口权限
				if IsOkPermissionsForAdmin(mysql.DB, ad.RoleId, c.Request.URL.Path, c.Query("action"), c.Request.RequestURI) == false {
					tools.JsonWrite(c, client.NoHavePermission, nil, "没有权限")
					c.Abort()
				}

				//设置who
				c.Set("who", ad)
				c.Next()
			} else {
				tools.JsonWrite(c, client.IllegalityCode, nil, client.IllegalityMsg)
				c.Abort()
			}

		}

		c.Next()

	}

}

// Cors 跨域设置
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		if origin != "" {
			//接收客户端发送的origin （重要！）
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			//服务器支持的所有跨域请求的方法
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			//允许跨域设置可以返回其他子段，可以自定义字段
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session")
			// 允许浏览器（客户端）可以解析的头部 （重要）
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			//设置缓存时间
			c.Header("Access-Control-Max-Age", "172800")
			//允许客户端传递校验信息比如 cookie (重要)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		//允许类型校验
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
		}

		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic info is: %v", err)
			}
		}()

		c.Next()
	}
}

// LimitIpRequestSameUrlForUser ip  访问限制(用户)
func LimitIpRequestSameUrlForUser() gin.HandlerFunc {
	return func(context *gin.Context) {
		//获取请求的地址
		urlPath := context.Request.URL.Path
		ip := context.ClientIP()
		if strings.Index(urlPath, "/client/") != -1 {
			//回去系统设置的ip限制次数
			var LimitTimes int
			rl, _ := redis.Rdb.HGet("SystemConfig", "RequestLimit").Result()
			if rl == "" {
				//读取数据库
				cf := model.Config{}
				err := mysql.DB.Where("id=?", 1).First(&cf).Error
				if err == nil {
					redis.Rdb.HMSet("SystemConfig", structs.Map(cf))
				} else {
					//写入默认值
					redis.Rdb.HSet("SystemConfig", "request_limit", 5) //1秒5次
				}

				LimitTimes = cf.RequestLimit
			}
			LimitTimes, _ = strconv.Atoi(rl)
			key := ip + urlPath
			curr := redis.Rdb.LLen(key).Val()
			if int(curr) >= LimitTimes {
				//超出了限制
				tools.JsonWrite(context, client.IpLimitWaring, nil, client.LimitWait)
				context.Abort()
			}
			if v := redis.Rdb.Exists(key).Val(); v == 0 {
				pipe := redis.Rdb.TxPipeline()
				pipe.RPush(key, key)
				//设置过期时间
				pipe.Expire(key, 1*time.Second)
				_, _ = pipe.Exec()
			} else {
				redis.Rdb.RPushX(key, key)
			}

			context.Next()
		}

	}
}

// IsOkPermissionsForAdmin 判断接口权限
func IsOkPermissionsForAdmin(db *gorm.DB, roleId int, path string, action string, RequestUrl string) bool {
	//获取role 权限字符串
	whiteUrl := []string{"/management/v1/system/country?action=select"}
	if tools.IsArray(whiteUrl, RequestUrl) {
		return true
	}

	ro := model.Role{}
	err := db.Where("id=?", roleId).First(&ro).Error
	if err != nil {
		return false
	}
	if ro.Jurisdiction == "" {
		return false
	}
	var data []model.RoleMenus
	err = json.Unmarshal([]byte(ro.Jurisdiction), &data)
	if err != nil {
		return false
	}
	arr := strings.Split(path, "/")
	if 2 <= len(arr) {
		path = arr[len(arr)-2] + "/" + arr[len(arr)-1]
	}
	for _, i := range data {
		for _, secondary := range i.Second {
			if secondary.RouterPath == path {
				if action == "add" && secondary.Add == 1 {
					return true
				}

				if action == "update" && secondary.Update == 1 {
					return true
				}

				if action == "delete" && secondary.Delete == 1 {
					return true
				}

				if action == "select" && secondary.Read == 1 {
					return true
				}

			}
		}

	}

	return false
}

package client

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/wangyi/GinTemplate/model"
	"github.com/wangyi/GinTemplate/tools"
	"net/http"
)

var (
	VerifyErrCode       = 1002  //校验错误Code
	ErrReturnCode       = -101  //全局返回错误的信息
	SuccessReturnCode   = 2000  //全局返回正常的数据
	IpLimitWaring       = -102  //全局限制警告
	IllegalityCode      = -103  //全局非法请求警告,一般是没有token
	NeedGoogleBind      = -104  //管理员登录需要绑定谷歌验证码
	TokenExpire         = -105  //管理员或者用户的token过期
	NoHavePermission    = -106  //没有权限
	MysqlErr            = -107  //数据的一些报错
	TaskClearing        = 300   //任务结算
	ReturnOldOrderCode  = 20001 //返回已经获取的账单
	NoBank              = 400
	SystemMinWithdrawal = 401
	NoEnoughMoney       = -888
)

var ()

// ReturnVerifyErrCode 参数非法的错误
func ReturnVerifyErrCode(context *gin.Context, err error) {
	context.JSON(http.StatusOK, gin.H{
		"code":   VerifyErrCode,
		"result": nil,
		"msg":    err.Error(),
	})
}

func ReturnErr101Code(context *gin.Context, msg interface{}) {
	context.JSON(http.StatusOK, gin.H{
		"code":   -101,
		"result": nil,
		"msg":    msg,
	})
}

func ReturnSuccess2000Code(context *gin.Context, msg string) {
	context.JSON(http.StatusOK, gin.H{
		"code":   SuccessReturnCode,
		"result": nil,
		"msg":    msg,
	})
}
func ReturnSuccess2000DataCode(context *gin.Context, data interface{}, msg string) {
	context.JSON(http.StatusOK, gin.H{
		"code":   SuccessReturnCode,
		"result": data,
		"msg":    msg,
	})
}

// CreateUserToken 生产 用户的Token
func CreateUserToken(db *gorm.DB) (string, bool) {
	for i := 0; i < 5; i++ {
		str := tools.RandStringRunes(36)
		//判断邀请码是否重复
		err := db.Where("token=?", str).First(&model.User{}).Error
		if err != nil {
			return str, true
		}
	}
	return "", false
}

func CreateUserInvitationCode(db *gorm.DB) (string, bool) {
	for i := 0; i < 5; i++ {
		str := tools.RandInvitationCode(6)
		//判断邀请码是否重复
		err := db.Where("invitation_code=?", str).First(&model.User{}).Error
		if err != nil {
			return str, true
		}
	}
	return "", false
}

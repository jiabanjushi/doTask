package client

// RegisterVerify 注册参数校验
type RegisterVerify struct {
	Username       string `form:"username"  binding:"required,min=4,max=10"`
	Phone          string `form:"phone"  binding:"omitempty,max=20"` //非必要
	Password       string `form:"password"  binding:"required,min=6,max=20"`
	PayPassword    string `form:"pay_password"  binding:"required,min=6,max=6"`
	InvitationCode string `form:"invitation_code"  binding:"required,min=8,max=8"` //邀请码
}

// LoginVerify 登录参数
type LoginVerify struct {
	Username string `form:"username"  binding:"required,min=4,max=10"`
	Password string `form:"password"  binding:"required,min=6,max=20"`
}

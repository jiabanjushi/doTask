package client

//注册返回信息
var (
	RegisterErr01   = "Sorry, the invitation code does not exist"                        //invitationNo
	RegisterErr02   = "Sorry, the user name already exists"                              //already
	RegisterErr03   = "Sorry, registration failed, system error, please try again later" //registration
	RegisterSuccess = "Congratulations, you have registered successfully"                //congratulations

	PhoneIsRegistered = "The cell phone number has been registered"
)

//登录返回信息

var (
	LoginErr01   = "Sorry, the account or password is wrong"        //loginPassword
	PasswordErr  = "The original password you entered is incorrect" //PasswordErr
	NoActivation = "NoActivation"
)

// LimitWait 全局作用
var (
	LimitWait     = "You clicked too fast. Please try again later" //LimitWait
	IllegalityMsg = "Sorry, your request is invalid"               //IllegalityMsg
	LoginExpire   = "Sorry, your login has expired"                //LoginExpire
)

//获取验证码返回错误信息

var (
	NoThisCountry = "There is no language of this country" //NoThisCountry
)

//  任务

var (
	TaskFrozen           = "Your assignment has been frozen. Please contact customer service"
	UnassignedTask       = "Sorry, you have no unassigned task"
	GetTaskErr           = "Description Failed to obtain a task."
	OnFindTaskOrderId    = "The order number does not exist"
	DonDoubleCommit      = "Don't double commit"
	DotEnoughMoney       = "Don't have enough money"
	ErrPayPassword       = "Wrong payment password"
	HaveTaskIsSettlement = "Some tasks are clearing up and will be obtained later"
	CantWithdraw         = "You can't withdraw money until it's finished"
)

//record

var (
	ChannelMaintenance = "Channel maintenance"
	NotRechargeMoney   = "The recharge amount can not be lower than or greater than the set value"
	NOEnoughMoney      = "Sorry, your balance is not enough"
	NoBindBankCard     = "Sorry, you don't have a bank card attached"
	PayFail            = "PayFail" //拉起支付订单失败
)

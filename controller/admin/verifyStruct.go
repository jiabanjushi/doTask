package admin

// LoginVerify 登录参数
type LoginVerify struct {
	Username   string `form:"adminUsername"  binding:"required"`
	Password   string `form:"adminPassword"  binding:"required"`
	GoogleCode string `form:"googleCode"  binding:"omitempty,max=6" `
}

//添加国家

type AddCountry struct {
	Country string `form:"country_name" binding:"required"`
}

// AddTaskVerify 上传任务
type AddTaskVerify struct {
	OverlayId         *int   `form:"overlay_id" binding:"required"`
	TaskName          string `form:"task_name" binding:"required"`
	TaskCount         int    `form:"task_count" binding:"omitempty"`
	AllCommissionRate string `form:"all_commission_rate" binding:"omitempty"`
	Dialog            int    `form:"dialog" binding:"required"`
	//DialogImage       string `form:"dialog_image"`
	VipId        int    `form:"vip_id" binding:"required"`
	OverlayIndex *int   `form:"overlay_index" binding:"required"`
	PayMod       string `form:"pay_mod" binding:"omitempty"`
}

//更新任务

type UpdateTaskVerify struct {
	TaskName          string `form:"task_name" binding:"omitempty"`
	AllCommissionRate string `form:"all_commission_rate" binding:"omitempty"`
	Dialog            int    `form:"dialog" binding:"omitempty"`
	VipId             int    `form:"vip_id" binding:"omitempty"`
	OverlayIndex      int    `form:"overlay_index" binding:"omitempty"`
	TaskCount         int    `form:"task_count" binding:"omitempty"`
	PayMod            string `form:"pay_mod" binding:"omitempty"`
}

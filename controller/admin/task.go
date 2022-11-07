package admin

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/common"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mmdb"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"strconv"
	"strings"
	"time"
)

// OperationTask 分组任务
func OperationTask(c *gin.Context) {

	action := c.Query("action")
	who, _ := c.Get("who")
	whoMap := who.(model.Admin)
	ctName, _ := mmdb.GetCountryForIp(c.ClientIP())
	if action == "add" {
		//判断创建任务的模式
		var av AddTaskVerify
		//检查参数
		if err := c.ShouldBind(&av); err != nil {
			client.ReturnVerifyErrCode(c, err)
			return
		}
		IfOverlay := false
		//判断 overlay_id
		if *av.OverlayId != 0 {
			//判断 父级任务是否存在
			err := mysql.DB.Where("id=? or overlay_index=?", *av.OverlayId, *av.OverlayIndex).First(&model.Task{}).Error
			if err != nil {
				client.ReturnErr101Code(c, "非法添加")
				return
			}
			IfOverlay = true

		}
		//检查 vip等级
		v := model.Vip{ID: av.VipId}
		if v.IsExist(mysql.DB) == false {
			client.ReturnErr101Code(c, "vip_id非法")
			return
		}
		ta := model.Task{
			TaskName:          av.TaskName,
			TaskCount:         av.TaskCount,
			VipId:             av.VipId,
			AllCommissionRate: av.AllCommissionRate,
			Dialog:            av.Dialog,
			OverlayIndex:      *av.OverlayIndex,
			OverlayId:         *av.OverlayId,
			IfOverlay:         IfOverlay,
			PayMod:            av.PayMod,
			TopAgent:          whoMap.AdminUser,
		}
		//检查是否有弹窗
		if av.Dialog == 2 {
			file, err := c.FormFile("dialog_image")
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}
			filepath := common.UploadTask + file.Filename
			err = c.SaveUploadedFile(file, filepath)
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}
			ta.DialogImage = filepath
		}

		err := ta.CreateNoOverLayTask(mysql.DB)
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}

		client.ReturnSuccess2000Code(c, "添加任务分组成功")
		//日志
		LO := model.Log{Country: ctName, Ip: c.ClientIP(), Content: whoMap.AdminUser + "|" + "创建任务分组,组名:" + av.TaskName, Kinds: 3}
		LO.CreateLogger(mysql.DB)
		return

	}

	if action == "select" {
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.Task, 0)
		db := mysql.DB
		var total int

		if whoMap.AgencyUsername != "" {
			db = db.Where("top_agent=?", whoMap.AdminUser)
		}

		if overlayId, isExist := c.GetPostForm("overlay_id"); isExist == true {
			db = db.Where("overlay_id=?", overlayId)

		}
		db.Model(model.Task{}).Count(&total)
		db = db.Model(&model.Task{}).Offset((page - 1) * limit).Limit(limit).Order("created asc")
		db.Find(&sl)
		for i, task := range sl {
			vp := model.Vip{}
			err := mysql.DB.Where("id=?", task.VipId).First(&vp).Error
			if err == nil {
				sl[i].VipName = vp.Name
			}
		}
		ReturnDataLIst2000(c, sl, total)
	}

	if action == "update" {
		id := c.PostForm("id")
		db := mysql.DB

		tt := model.Task{}
		err := db.Where("id=?", id).First(&tt).Error
		if err != nil {
			client.ReturnErr101Code(c, "修改的任务不存在")
			return
		}
		var up UpdateTaskVerify
		//检查参数
		if err := c.ShouldBind(&up); err != nil {
			client.ReturnVerifyErrCode(c, err)
			return
		}

		if tt.OverlayId != 0 {
			ta := model.Task{}
			err := db.Where("overlay_id=?   and overlay_index=?", tt.OverlayId, up.OverlayIndex).First(&ta).Error
			idInt, _ := strconv.Atoi(id)
			if err == nil && ta.ID != idInt {
				client.ReturnErr101Code(c, "二级任务overlay_index不可以一样")
				return
			}
		}

		new := model.Task{}
		new.TaskName = up.TaskName
		new.VipId = up.VipId
		new.OverlayIndex = up.OverlayIndex
		new.AllCommissionRate = up.AllCommissionRate
		new.TaskCount = up.TaskCount
		new.PayMod = up.PayMod
		new.Dialog = up.Dialog
		if up.Dialog == 2 {
			file, err := c.FormFile("dialog_image")
			if err == nil {
				filepath := common.UploadTask + file.Filename
				err = c.SaveUploadedFile(file, filepath)
				if err != nil {
					client.ReturnErr101Code(c, err.Error())
					return
				}
				new.DialogImage = filepath
			}
		}

		err = db.Model(&model.Task{}).Where("id=?", id).Update(&new).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "修改成功")
		//日志

		marshal, err := json.Marshal(new)
		if err == nil {
			LO := model.Log{Country: ctName, Ip: c.ClientIP(), Content: whoMap.AdminUser + "|" + "修改任务,id:" + id + " 内容:" + string(marshal), Kinds: 3}
			LO.CreateLogger(mysql.DB)
		}
		return
	}

	if action == "delete" {
		id := c.PostForm("id")
		db := mysql.DB
		a := model.Task{}
		err := db.Where("id=?", id).First(&a).Error
		if err != nil {
			client.ReturnErr101Code(c, "修改的任务不存在")
			return
		}

		//判断删除的叠加任务 还是 小的任务
		if a.OverlayId == 0 && a.OverlayIndex == 0 {
			db = db.Begin()
			err := db.Where("id=?", id).Delete(&model.Task{}).Error
			err1 := db.Where("overlay_id=?", id).Delete(&model.Task{}).Error
			if err != nil || err1 != nil {
				db.Rollback()
				client.ReturnErr101Code(c, err.Error())
				return
			}
			db.Commit()
		} else {
			err := db.Delete(&model.Task{}, id).Error
			if err != nil {
				client.ReturnErr101Code(c, err.Error())
				return
			}

		}
		client.ReturnSuccess2000Code(c, "删除成功")
		return

	}
}

// OperationTaskOder 任务订单操作
func OperationTaskOder(c *gin.Context) {
	action := c.Query("action")
	who, _ := c.Get("who")
	whoMap := who.(model.Admin)
	//ctName, _ := mmdb.GetCountryForIp(c.ClientIP())

	if action == "select" {
		limit, _ := strconv.Atoi(c.PostForm("limit"))
		page, _ := strconv.Atoi(c.PostForm("page"))
		sl := make([]model.TaskOrder, 0)
		db := mysql.DB
		var total int

		if whoMap.AgencyUsername != "" {
			var p []int
			arrayUg := strings.Split(whoMap.AgencyUsername, ",")
			for _, s := range arrayUg {
				us := make([]model.User, 0)
				mysql.DB.Where("top_agent=?", s).Find(&us)
				for _, u := range us {
					p = append(p, u.ID)
				}
			}
			db = db.Where("uid  in (?)", p)
		}

		if overlayId, isExist := c.GetPostForm("username"); isExist == true {
			if overlayId != "" {
				user := model.User{}
				mysql.DB.Where("username=?", overlayId).First(&user)
				db = db.Where("uid=?", user.ID)
			}
		}

		if overlayId, isExist := c.GetPostForm("status"); isExist == true {
			db = db.Where("status=?", overlayId)
		}

		db.Model(model.TaskOrder{}).Count(&total)
		db = db.Model(&model.TaskOrder{}).Offset((page - 1) * limit).Limit(limit).Order("created asc")
		db.Find(&sl)
		for i, order := range sl {
			good := model.Goods{}
			mysql.DB.Where("id=?", order.GoodsId).First(&good)
			sl[i].GoodsName = good.GoodsName
			sl[i].GoodsImageUrl = good.GoodsImages
			user := model.User{}
			mysql.DB.Where("id=?", order.Uid).First(&user)
			sl[i].Username = user.Username
			task := model.Task{}
			mysql.DB.Where("id=?", order.TaskId).First(&task)
			sl[i].TaskName = task.TaskName
			sl[i].TopAgent = user.TopAgent
		}

		ReturnDataLIst2000(c, sl, total)
	}

	if action == "update" {
		id := c.PostForm("id")
		taskOrder := model.TaskOrder{}
		err := mysql.DB.Where("id=?", id).First(&taskOrder).Error
		if err != nil {
			client.ReturnErr101Code(c, "任务订单不存在")
			return
		}
		if taskOrder.Status != 6 {
			client.ReturnErr101Code(c, "状态错误!")
			return
		}
		config := model.Config{}
		mysql.DB.Where("id=?", 1).First(&config)
		err = mysql.DB.Model(&model.TaskOrder{}).Where("id=?", id).Update(&model.TaskOrder{Status: 2, GetAt: time.Now().Unix() + config.TaskTimeout}).Error
		if err != nil {
			client.ReturnErr101Code(c, err.Error())
			return
		}
		client.ReturnSuccess2000Code(c, "解除成功")
		return

	}

}

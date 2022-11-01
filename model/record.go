package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/wangyi/GinTemplate/tools"
	"strconv"
	"time"
)

type Record struct {
	ID                int    `gorm:"primaryKey"`
	OrderNum          string //账单单号  任务佣金开头 YJ
	Kinds             int    //1 提现  2 充值 3任务佣金 4人工上分
	Status            int
	RejectReason      string //  针对提现 驳回原因
	PayFailReason     string //针对三方代付失败原因
	Remark            string
	Money             float64 `gorm:"type:decimal(10,2)"` //涉及金额
	TaskOrderId       int     //涉及佣金
	Created           int64
	Updated           int64
	UserId            int
	Date              string //日期
	Artificial        int    `gorm:"default:1"` //人工加框  状态1 不是 2是
	PayChannelsId     int    //充值渠道id
	ServiceCharge     float64
	OnLine            int         //1线上  2线下
	AuthenticityMoney float64     `gorm:"type:decimal(10,2);default:0.00"` //真实金额
	SystemMoney       float64     `gorm:"type:decimal(10,2);default:0.00"` //系统金额
	PayChannel        PayChannels `gorm:"-"`
	UserName          string      `gorm:"-"`

	TopAgent string `gorm:"-"`
}

// CheckIsExistModelRecord   创建User
func CheckIsExistModelRecord(db *gorm.DB) {
	if db.HasTable(&Record{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&Record{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&Record{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

func (r *Record) CreatedRecord(db *gorm.DB) (int, error) {
	r.Created = time.Now().Unix()
	r.Updated = time.Now().Unix()
	r.Date = time.Now().Format("2006-01-02")
	if r.Kinds == 3 {
		r.OrderNum = "YJ" + tools.RandStringRunesZM(2) + time.Now().Format("20060102150405") + strconv.Itoa(int(tools.GetRandomWithAll(100, 999)))
		err := db.Save(r).Error
		if err != nil {
			return 0, err
		}
	}

	//提现
	if r.Kinds == 1 {
		r.OrderNum = "TX" + tools.RandStringRunesZM(2) + time.Now().Format("20060102150405") + strconv.Itoa(int(tools.GetRandomWithAll(100, 999)))
		err := db.Save(r).Error
		if err != nil {
			return 0, err
		}
	}
	return r.ID, nil
}

// CreatedRechargeOrder 生成充值订单
func (r *Record) CreatedRechargeOrder(db *gorm.DB) (*Record, error) {
	r.Created = time.Now().Unix()
	r.Updated = time.Now().Unix()
	r.Status = 1
	r.Kinds = 2
	r.OrderNum = "CZ" + tools.RandStringRunesZM(2) + time.Now().Format("20060102150405") + strconv.Itoa(int(tools.GetRandomWithAll(100, 999)))

	err := db.Save(&r).Error
	if err != nil {
		return r, err
	}
	return r, nil
}

// UpdateRecordToPoint 上分
func (r *Record) UpdateRecordToPoint(db *gorm.DB) (bool, error) {
	r.Updated = time.Now().Unix()
	r.Status = 3
	r.Date = time.Now().Format("2006-01-02")
	err := db.Model(&Record{}).Where("id=?", r.ID).Update(r).Error
	if err != nil {
		return false, err
	}

	//充值上分   这个接口一般用于
	return true, nil
}

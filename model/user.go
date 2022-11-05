package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/wangyi/GinTemplate/common"
	eeor "github.com/wangyi/GinTemplate/error"
	"math"
	"sync"
	"time"
)

//var globalUserMap = sync.Map{}
//
//func GetUserLock(UserId int) sync.RWMutex {
//	lok, _ := globalUserMap.LoadOrStore(UserId, sync.RWMutex{})
//	fmt.Println(lok)
//	return lok.(sync.RWMutex)
//}

// User 用户
type User struct {
	ID               int `gorm:"primaryKey"`
	Username         string
	Password         string
	Token            string
	InvitationCode   string
	PayPassword      string
	Phone            string
	LevelTree        string  `gorm:"comment:'等级树,分号进行分割'"`
	TopAgent         string  `gorm:"comment:'顶级代理'"`
	SuperiorAgent    string  `gorm:"comment:'上级代理'"`
	Balance          float64 `gorm:"type:decimal(10,2);comment:'账户余额' ;default:0"`
	WithdrawFreeze   float64 `gorm:"type:decimal(10,2);comment:'提现冻结金额';default:0"`
	WorkingFreeze    float64 `gorm:"type:decimal(10,2);comment:'任务冻结金额';default:0"`
	VipId            int     `gorm:"default:1"`
	Status           int     `gorm:"comment: '状态  1正常  2禁用';default:1"`
	Kinds            int     `gorm:"comment:'账号类型 1 正式号 2 测试号';default:1"`
	Created          int64   `gorm:"comment:'用户注册的时间'"`
	CreatedIp        string  `gorm:"comment:'注册ip'"`
	CreatedCountry   string  `gorm:"comment:'注册的国家'"`
	TheScLoginIp     string  `gorm:"comment:'上次登录ip'"`
	TheScLoginTime   int64   `gorm:"comment:'上次登录时间'"`
	TheLastLoginTime int64   `gorm:"comment:'最后一次登录时间'"`
	TheLastLoginIp   string  `gorm:"comment:'最后一次登录ip'"`
	//其他参数便于前段显示的
	VipName  string `gorm:"-"`
	UserLock sync.RWMutex
	//顶级会员应该拥有的数据
	NumberNum int         `gorm:"-"` //成员个数
	DoingTask Task        `gorm:"-"` //成员个数
	Extend    interface{} `gorm:"-"` //扩展
}

// CheckIsExistModelUser 创建User
func CheckIsExistModelUser(db *gorm.DB) {
	if db.HasTable(&User{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&User{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&User{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

// CreateUser 创建用户
func (u *User) CreateUser(db *gorm.DB) {

}

// GetInvitationCode 通过邀请码 获取用户
func (u *User) GetInvitationCode(db *gorm.DB) {
	err := db.Where("invitation_code=?", u.InvitationCode).First(&u).Error
	if err != nil {
		//邀请码不存在
		return
	}

}

type UserBalanceChange struct {
	UserId                  int
	TaskOrderId             int
	ChangeMoney             float64 //变化的金额
	Kinds                   int     // 1 提交任务订单()  2结算任务订单  3人工加减余额  4提现     5 充值成功(用户已上分,并且产生账变)   6提现被驳回()    7代付成功
	RecordKind              int     //账单类型
	RecordId                int     //账单id  (上分)
	AuthenticityMoney       float64 //充值 真实金额
	PayChannelsExchangeRate float64 //充值 渠道的会汇率
	RejectReason            string  //针对提现 驳回原因
	OrderNo                 string  //三方订单
	PaymentTime             string  //支付时间
}

// UserBalanceChangeFunc 全局 余额 操作
func (Ubc *UserBalanceChange) UserBalanceChangeFunc(db *gorm.DB) (int, error) {
	common.LockForGlobalChangeBalance.RLock() //读锁
	//读取用户的  余额
	user := User{}
	err := db.Where("id=?", Ubc.UserId).First(&user).Error
	if err != nil {
		common.LockForGlobalChangeBalance.RUnlock()
		return -2, err //没有找到这个用户
	}
	if Ubc.ChangeMoney < 0 {
		//余额 扣除操作
		if math.Abs(Ubc.ChangeMoney) > user.Balance {
			//用户的余额不足
			common.LockForGlobalChangeBalance.RUnlock()
			if Ubc.Kinds == 1 {
				return -1, eeor.OtherError(fmt.Sprintf("%f", math.Abs(Ubc.ChangeMoney-user.Balance)))
			}
			return -1, eeor.OtherError("Don't have enough money")
		}
	}
	common.LockForGlobalChangeBalance.RUnlock() //解锁
	//上写锁
	common.LockForGlobalChangeBalance.Lock()
	defer common.LockForGlobalChangeBalance.Unlock() //解除读锁
	db = db.Begin()
	if Ubc.Kinds == 1 {
		//更新账户的余额
		newUser := map[string]interface{}{}
		newUser["Balance"] = user.Balance + Ubc.ChangeMoney
		newUser["WorkingFreeze"] = user.WorkingFreeze + math.Abs(Ubc.ChangeMoney)
		err := db.Model(&User{}).Where("id=?", Ubc.UserId).Update(newUser).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}
		//生成账变记录
		changeMoney := AccountChange{
			Created:       time.Now().Unix(),
			UserId:        Ubc.UserId,
			Money:         Ubc.ChangeMoney,
			NowMoney:      user.Balance + Ubc.ChangeMoney,
			OriginalMoney: user.Balance,
			TaskOrderId:   Ubc.TaskOrderId,
		}
		err = db.Save(&changeMoney).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}
		//修改任务订单的状态
		err = db.Model(&TaskOrder{}).Where("id=?", Ubc.TaskOrderId).Update(&TaskOrder{
			Updated: time.Now().Unix(),
			Status:  5,
		}).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}

	}
	if Ubc.Kinds == 2 {
		taskOrder := TaskOrder{}
		err := db.Where("id=?", Ubc.TaskOrderId).First(&taskOrder).Error
		if err != nil {
			return -2, err //没有找到这个任务订单
		}
		//首先生成佣金订单
		newUser := map[string]interface{}{}
		var NowMoney float64
		var brokerage float64 //佣金金额
		if taskOrder.CommissionMoney != 0 {
			brokerage = taskOrder.CommissionMoney
		}
		if taskOrder.CommissionRate != 0 {
			brokerage = taskOrder.OrderMoney * taskOrder.CommissionRate //订单的金额 *  佣金比例
		}
		NowMoney = brokerage + user.Balance
		newUser["Balance"] = NowMoney + taskOrder.OrderMoney
		newUser["WorkingFreeze"] = user.WorkingFreeze - taskOrder.OrderMoney
		//修改账户的余额
		err = db.Model(&User{}).Where("id=?", Ubc.UserId).Update(newUser).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}
		//生成佣金订单
		record := Record{Kinds: 3, Money: brokerage, Status: 1, TaskOrderId: Ubc.TaskOrderId, UserId: Ubc.UserId}
		RId, err := record.CreatedRecord(db)
		if err != nil {
			db.Rollback()
			return -2, err
		}
		//生产账变订单
		Cm := AccountChange{
			UserId:        Ubc.UserId,
			Money:         brokerage,
			OriginalMoney: user.Balance,
			NowMoney:      NowMoney,
			RecordId:      RId,
		}
		err = db.Save(&Cm).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}
		//解冻 任务金额  生产账变订单
		Cm1 := AccountChange{
			UserId:        Ubc.UserId,
			Money:         taskOrder.OrderMoney,
			OriginalMoney: NowMoney,
			NowMoney:      NowMoney + taskOrder.OrderMoney,
			TaskOrderId:   taskOrder.ID,
		}
		err = db.Save(&Cm1).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}
		//修改任务订单的状态  由5已提交,未结算 -> 3已完成
		err = db.Model(&TaskOrder{}).Where("id=?", taskOrder.ID).Update(&TaskOrder{Updated: time.Now().Unix(), Status: 3}).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}

	}
	//3人工加减余额
	if Ubc.Kinds == 3 {
		newUser := map[string]interface{}{}
		newUser["Balance"] = user.Balance + Ubc.ChangeMoney
		err := db.Model(&User{}).Where("id=?", Ubc.UserId).Update(newUser).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}
		//生成账变记录
		changeMoney := AccountChange{
			Created:       time.Now().Unix(),
			UserId:        Ubc.UserId,
			Money:         Ubc.ChangeMoney,
			NowMoney:      user.Balance + Ubc.ChangeMoney,
			OriginalMoney: user.Balance,
			TaskOrderId:   Ubc.TaskOrderId,
		}
		err = db.Save(&changeMoney).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}

		//生成record
		record := Record{Kinds: Ubc.RecordKind, Money: Ubc.ChangeMoney, Status: 1, UserId: Ubc.UserId, Artificial: 2}
		_, err = record.CreatedRecord(db)
		if err != nil {
			db.Rollback()
			return -2, err
		}

	}
	//4提现
	if Ubc.Kinds == 4 {
		//更新账户的余额
		newUser := map[string]interface{}{}
		newUser["Balance"] = user.Balance + Ubc.ChangeMoney
		newUser["WithdrawFreeze"] = user.WorkingFreeze + math.Abs(Ubc.ChangeMoney)
		err := db.Model(&User{}).Where("id=?", Ubc.UserId).Update(newUser).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}

		config := Config{}
		db.Where("id=?", 1).First(&config)
		//生成 提现订单
		re := Record{Kinds: 1, UserId: Ubc.UserId, Status: 1, Money: math.Abs(Ubc.ChangeMoney), ServiceCharge: config.WithdrawalHand}
		RecordId, err := re.CreatedRecord(db)
		if err != nil {
			db.Rollback()
			return -2, err
		}

		//账变账单
		AC := AccountChange{UserId: Ubc.UserId, RecordId: RecordId, Money: math.Abs(Ubc.ChangeMoney), OriginalMoney: user.Balance, NowMoney: user.Balance + Ubc.ChangeMoney}
		_, err = AC.CreatedAccountChange(db)
		if err != nil {
			db.Rollback()
			return -2, err
		}

	}
	//5充值成功
	if Ubc.Kinds == 5 {
		newUser := map[string]interface{}{}
		newUser["Balance"] = user.Balance + Ubc.AuthenticityMoney*Ubc.PayChannelsExchangeRate
		err := db.Model(&User{}).Where("id=?", Ubc.UserId).Update(newUser).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}

		//账单状态修改
		re := Record{ID: Ubc.RecordId,
			AuthenticityMoney: Ubc.AuthenticityMoney,
			SystemMoney:       Ubc.AuthenticityMoney * Ubc.PayChannelsExchangeRate,
			ThreeOrderNum:     Ubc.OrderNo,
			PaymentTime:       Ubc.PaymentTime,
		}
		_, err = re.UpdateRecordToPoint(db)
		if err != nil {
			db.Rollback()
			return -2, err
		}

		//账变
		AC := AccountChange{UserId: Ubc.UserId, Money: Ubc.AuthenticityMoney * Ubc.PayChannelsExchangeRate, OriginalMoney: user.Balance, NowMoney: user.Balance + Ubc.AuthenticityMoney*Ubc.PayChannelsExchangeRate, RecordId: Ubc.RecordId}
		_, err = AC.CreatedAccountChange(db)
		if err != nil {
			db.Rollback()
			return -2, err
		}

	}
	//6提现被驳回
	if Ubc.Kinds == 6 {
		//修改玩家余额
		newUser := map[string]interface{}{}
		newUser["Balance"] = user.Balance + Ubc.ChangeMoney
		newUser["WithdrawFreeze"] = user.WithdrawFreeze - Ubc.ChangeMoney
		fmt.Println(newUser["WithdrawFreeze"])

		err := db.Model(&User{}).Where("id=?", Ubc.UserId).Update(newUser).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}
		//生成账变订单
		AC := AccountChange{
			UserId:        Ubc.UserId,
			RecordId:      Ubc.RecordId,
			Money:         math.Abs(Ubc.ChangeMoney),
			OriginalMoney: user.Balance,
			NowMoney:      user.Balance + Ubc.ChangeMoney}
		_, err = AC.CreatedAccountChange(db)
		if err != nil {
			db.Rollback()
			return -2, err
		}
		//修改订单
		err = db.Model(&Record{}).Where("id=?", Ubc.RecordId).Update(&Record{Status: 6, Updated: time.Now().Unix(), RejectReason: Ubc.RejectReason}).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}

	}
	//7代付成功
	if Ubc.Kinds == 7 {
		updateMap := make(map[string]interface{})
		updateMap["WithdrawFreeze"] = user.WithdrawFreeze - Ubc.ChangeMoney
		err = db.Model(&User{}).Where("id=?", user.ID).Update(updateMap).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}
		err = db.Model(&Record{}).Where("id=?", Ubc.RecordId).Update(&Record{Status: 5, Updated: time.Now().Unix(), Date: time.Now().Format("2006-01-02")}).Error
		if err != nil {
			db.Rollback()
			return -2, err
		}
	}

	db.Commit()
	return 1, nil
}

func GetRealBalance(db *gorm.DB, userId int) float64 {
	u := User{}
	db.Where("id=?", userId).First(&u)
	return u.Balance + u.WorkingFreeze
}

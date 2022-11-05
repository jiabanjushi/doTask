package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	eeor "github.com/wangyi/GinTemplate/error"
	"github.com/wangyi/GinTemplate/pay"
	"strconv"
)

// PayChannels 支付/代付通道
type PayChannels struct {
	ID             int `gorm:"primaryKey"`
	Kinds          int
	PayUrl         string //支付地址
	CountryId      int
	CurrencySymbol string //货币符号
	OnLine         int
	BackUrl        string
	Merchants      string //商户号
	PayCode        string //支付代码
	Key            string //支付/代付秘钥
	PayInterval    string //支付区间
	BackIp         string //back_ip
	BankPayId      int
	Created        int64
	Updated        int64
	Maintenance    int
	Status         int
	Name           string //名字
	PayType        int    //1  USDT   2BPay(支付)
	CountryCode    string //国际代码(BPay)
	Goods          string
	ExtendedParams string //代付的扩展参数
	PayFast        string
	PublicKey      string  `gorm:"type:text"`
	PrivateKey     string  `gorm:"type:text"`
	ExchangeRate   float64 `gorm:"type:decimal(10,2)"`
	CountryName    string  `gorm:"-"`
	BankPayIDName  string  `gorm:"-"`
}

type PayChannelsChoose struct {
	PayChannels PayChannels
	Record      Record
}

func CheckIsExistModelPayChannels(db *gorm.DB) {
	if db.HasTable(&PayChannels{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&PayChannels{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&PayChannels{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

// ChoosePay 选择支付
func (py *PayChannelsChoose) ChoosePay(db *gorm.DB) (string, error) {
	//墨西哥BPay
	if py.PayChannels.PayType == 2 {
		//BPay
		pb := pay.BPay{
			MerchantNo:      py.PayChannels.Merchants,
			MerchantOrderNo: py.Record.OrderNum,
			CountryCode:     py.PayChannels.CountryCode,
			CurrencyCode:    py.PayChannels.CurrencySymbol,
			PaymentType:     py.PayChannels.PayCode,
			PaymentAmount:   strconv.FormatFloat(py.Record.Money, 'f', 2, 64),
			Goods:           py.PayChannels.Goods,
			NotifyUrl:       py.PayChannels.BackUrl,
			PayUrl:          py.PayChannels.PayUrl,
			PrivateKey:      py.PayChannels.PrivateKey,
			PublicKey:       py.PayChannels.PublicKey,
		}
		order, err := pb.CreatedOrder(db)
		if err != nil {
			return "", err
		}
		return order, nil
	}

	return "", eeor.OtherError("There is no matching PayType ")
}

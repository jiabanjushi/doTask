package pay

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"github.com/wangyi/GinTemplate/pay"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type BackPay struct {
	OrderNo         string `json:"orderNo"`
	OrderTime       string `json:"orderTime"`
	OrderAmount     string `json:"orderAmount"`
	CountryCode     string `json:"countryCode"`
	Sign            string `json:"sign"`
	PaymentTime     string `json:"paymentTime"`
	MerchantOrderNo string `json:"merchantOrderNo"`
	PaymentAmount   string `json:"paymentAmount"`
	CurrencyCode    string `json:"currencyCode"`
	PaymentStatus   string `json:"paymentStatus"`
	//ReturnedParams  string `json:"returnedParams"`
	MerchantNo string `json:"merchantNo"`
}

// BackPayBPay BackBPay BPay  支付 回调
func BackPayBPay(c *gin.Context) {
	var bp BackPay
	if err := c.BindJSON(&bp); err != nil {
		client.ReturnErr101Code(c, err.Error())
		return
	}
	//Bpay
	pc := model.PayChannels{}
	err := mysql.DB.Where("pay_type=? and kinds=?", 2, 1).First(&pc).Error
	if err != nil {
		zap.L().Debug("pay|BackPayBPay|error:" + err.Error())
		client.ReturnErr101Code(c, err.Error())
		return
	}

	//校验签名
	signStr := "countryCode=" + bp.CountryCode + "&currencyCode=" + bp.CurrencyCode + "&merchantNo=" + bp.MerchantNo + "&merchantOrderNo=" + bp.MerchantOrderNo + "&orderAmount=" + bp.OrderAmount + "&orderNo=" + bp.OrderNo + "&orderTime=" + bp.OrderTime + "&paymentAmount=" + bp.PaymentAmount + "&paymentStatus=" + bp.PaymentStatus + "&paymentTime=" + bp.PaymentTime
	fmt.Println(signStr)
	_, err = pay.VerifyRsaSign(signStr, bp.Sign, pc.PublicKey)
	if err != nil {
		zap.L().Debug("pay|BackPayBPay|校验签名失败哦:" + err.Error())
		client.ReturnErr101Code(c, err.Error())
		return
	}

	//检验成功
	//查询订单
	if bp.PaymentStatus != "SUCCESS" {
		client.ReturnErr101Code(c, "不接收失败的订单")
		return
	}

	record := model.Record{}
	err = mysql.DB.Where("order_num=?", bp.MerchantOrderNo).First(&record).Error
	if err != nil {
		zap.L().Debug("pay|BackPayBPay|订单:" + bp.MerchantOrderNo + ",不存在")
		client.ReturnErr101Code(c, "无效订单号")
		return
	}
	//订单待支付  //订单超时
	if record.Status != 1 && record.Status != 4 {
		client.ReturnErr101Code(c, "不要重复提交")
		return
	}
	//查看设置
	config := model.Config{}
	mysql.DB.Where("id=?", 1).First(&config)
	if config.AutomaticPoints == 2 {
		//不自动上分
		AuthenticityMoney, _ := strconv.ParseFloat(bp.PaymentAmount, 64)
		SystemMoney := AuthenticityMoney * pc.ExchangeRate
		mysql.DB.Model(&model.Record{}).Where("id=?", record.ID).Update(&model.Record{
			Updated:           time.Now().Unix(),
			Status:            2,
			ThreeOrderNum:     bp.OrderNo,
			PaymentTime:       bp.PaymentTime,
			AuthenticityMoney: AuthenticityMoney, SystemMoney: SystemMoney, Date: time.Now().Format("2006-01-02")})
		c.String(http.StatusOK, "SUCCESS")
		return
	}

	AuthenticityMoney, _ := strconv.ParseFloat(bp.PaymentAmount, 64)
	//直接上分
	change := model.UserBalanceChange{
		UserId:                  record.UserId,
		Kinds:                   5,
		RecordId:                record.ID,
		PayChannelsExchangeRate: pc.ExchangeRate,
		AuthenticityMoney:       AuthenticityMoney,
		OrderNo:                 bp.OrderNo,
		PaymentTime:             bp.PaymentTime,
	}
	_, err = change.UserBalanceChangeFunc(mysql.DB)
	if err != nil {
		zap.L().Debug("pay|BackPayBPay|订单:" + bp.MerchantOrderNo + ",错误:" + err.Error())
		client.ReturnErr101Code(c, err.Error())
		return
	}
	c.String(http.StatusOK, "SUCCESS")
	return

}

type BackPaid struct {
	OrderNo         string `json:"orderNo"`
	OrderTime       string `json:"orderTime"`
	TransferStatus  string `json:"transferStatus"`
	TransferTime    string `json:"transferTime"`
	CountryCode     string `json:"countryCode"`
	OrderAmout      string `json:"orderAmout"`
	TransferAmount  string `json:"transferAmount"`
	Sign            string `json:"sign"`
	MerchantOrderNo string `json:"merchantOrderNo"`
	CurrencyCode    string `json:"currencyCode"`
	MerchantNo      string `json:"merchantNo"`
}

func BackPaidBPay(c *gin.Context) {
	var bp BackPaid
	if err := c.BindJSON(&bp); err != nil {
		client.ReturnSuccess2000Code(c, err.Error())
		return
	}
	//Bpay
	pc := model.PayChannels{}
	err := mysql.DB.Where("pay_type=? and kinds=?", 2, 2).First(&pc).Error
	if err != nil {
		zap.L().Debug("pay|BackPaidBPay|error:" + err.Error())
		client.ReturnSuccess2000Code(c, err.Error())
		return
	}

	//校验签名
	signStr := "countryCode=" + bp.CountryCode + "&currencyCode=" + bp.CurrencyCode + "&merchantNo=" + bp.MerchantNo + "&merchantOrderNo=" + bp.MerchantOrderNo + "&orderAmout=" + bp.OrderAmout + "&orderNo=" + bp.OrderNo + "&orderTime=" + bp.OrderTime + "&transferAmount=" + bp.TransferAmount + "&transferStatus=" + bp.TransferStatus + "&transferTime=" + bp.TransferTime
	zap.L().Debug("pay|BackPaidBPay|签名的字符串:" + signStr)
	_, err = pay.VerifyRsaSign(signStr, bp.Sign, pc.PublicKey)
	if err != nil {
		zap.L().Debug("pay|BackPaidBPay|校验签名失败哦:" + err.Error())
		client.ReturnSuccess2000Code(c, err.Error())
		return
	}
	//签名成功
	record := model.Record{}
	err = mysql.DB.Where("order_num=?", bp.MerchantOrderNo).First(&record).Error
	if err != nil {
		zap.L().Debug("pay|BackPaidBPay|订单:" + bp.MerchantOrderNo + ",不存在")
		client.ReturnErr101Code(c, "无效订单号")
		return
	}

	if record.Status == 5 {
		c.String(http.StatusOK, "SUCCESS")
		return
	}

	if bp.TransferStatus != "SUCCESS" {
		mysql.DB.Model(&model.Record{}).Where("id=?", record.ID).Update(&model.Record{Status: 4, PayFailReason: bp.TransferStatus, Updated: time.Now().Unix()})
		c.String(http.StatusOK, "SUCCESS")
		return
	}

	//回调成功
	c.String(http.StatusOK, "SUCCESS")
	db := mysql.DB.Begin()
	float, err := strconv.ParseFloat(bp.TransferAmount, 64)
	err = db.Model(&model.Record{}).Where("id=?", record.ID).Update(&model.Record{
		Status:            5,
		Updated:           time.Now().Unix(),
		AuthenticityMoney: float,
		ThreeOrderNum:     bp.OrderNo,
		PaymentTime:       bp.OrderTime,
	}).Error
	if err != nil {
		db.Rollback()
		zap.L().Debug("pay|BackPaidBPay|175|订单:" + bp.MerchantOrderNo + ",err:" + err.Error())
		return
	}
	//修改冻结提现金额
	user := model.User{}
	err = db.Where("id=?", record.UserId).First(&user).Error
	if err != nil {
		db.Rollback()
		zap.L().Debug("pay|BackPaidBPay|182|订单:" + bp.MerchantOrderNo + ",err:" + err.Error())
		return
	}

	err = db.Model(&model.User{}).Where("id=?", record.UserId).Update(map[string]interface{}{"WithdrawFreeze": user.WithdrawFreeze - record.Money}).Error
	if err != nil {
		db.Rollback()
		zap.L().Debug("pay|BackPaidBPay|192|订单:" + bp.MerchantOrderNo + ",err:" + err.Error())
		return
	}

	db.Commit()
	return

}

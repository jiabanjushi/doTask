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
		client.ReturnSuccess2000Code(c, err.Error())
		return
	}
	//Bpay
	pc := model.PayChannels{}
	err := mysql.DB.Where("pay_type=?", 2).First(&pc).Error
	if err != nil {
		zap.L().Debug("pay|BackPayBPay|error:" + err.Error())
		client.ReturnSuccess2000Code(c, err.Error())

		return
	}

	//校验签名
	signStr := "countryCode=" + bp.CountryCode + "&currencyCode=" + bp.CurrencyCode + "&merchantNo=" + bp.MerchantNo + "&merchantOrderNo=" + bp.MerchantOrderNo + "&orderAmount=" + bp.OrderAmount + "&orderNo=" + bp.OrderNo + "&orderTime=" + bp.OrderTime + "&paymentAmount=" + bp.PaymentAmount + "&paymentStatus=" + bp.PaymentStatus + "&paymentTime=" + bp.PaymentTime
	fmt.Println(signStr)
	_, err = pay.VerifyRsaSign(signStr, bp.Sign, pc.PublicKey)
	if err != nil {
		zap.L().Debug("pay|BackPayBPay|校验签名失败哦:" + err.Error())
		client.ReturnSuccess2000Code(c, err.Error())
		return
	}

	//检验成功
	//查询订单

	if bp.PaymentStatus != "SUCCESS" {
		client.ReturnSuccess2000Code(c, "不接收失败的订单")
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
	change := model.UserBalanceChange{UserId: record.UserId, Kinds: 5, RecordId: record.ID, PayChannelsExchangeRate: pc.ExchangeRate, AuthenticityMoney: AuthenticityMoney}
	_, err = change.UserBalanceChangeFunc(mysql.DB)
	if err != nil {
		zap.L().Debug("pay|BackPayBPay|订单:" + bp.MerchantOrderNo + ",错误:" + err.Error())
		client.ReturnErr101Code(c, err.Error())
		return
	}
	c.String(http.StatusOK, "SUCCESS")
	return

}

func BackPaidBPay(c *gin.Context) {

}

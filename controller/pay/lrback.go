package pay

import (
	"github.com/gin-gonic/gin"
	"github.com/wangyi/GinTemplate/controller/client"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/model"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type BackPayLrPayData struct {
	BusiCode    string `json:"busi_code"`    //支付类型编码
	ErrCode     string `json:"err_code"`     //错误码
	ErrMsg      string `json:"err_msg"`      //错误描述
	MerNo       string `json:"mer_no"`       //商户唯一订单号
	MerOrderNo  string `json:"mer_order_no"` //商户订单号
	OrderAmount string `json:"order_amount"` //订单金额
	OrderNo     string `json:"order_no"`     //平台订单号
	OrderTime   string `json:"order_time"`   //订单时间
	PayAmount   string `json:"pay_amount"`   //支付金额
	PayTime     string `json:"pay_time"`     //支付时间
	Status      string `json:"status"`       //订单状态
	Sign        string `json:"sign"`         //数字签名

}

// BackPayLrPay lr 支付回调接口
func BackPayLrPay(c *gin.Context) {
	var bp BackPayLrPayData
	if err := c.BindJSON(&bp); err != nil {
		client.ReturnSuccess2000Code(c, err.Error())
		return
	}

	pc := model.PayChannels{}
	err := mysql.DB.Where("pay_type=? and kinds=?", 3, 1).First(&pc).Error
	if err != nil {
		zap.L().Debug("pay|BackPayLrPay|error:" + err.Error())
		client.ReturnSuccess2000Code(c, err.Error())
		return
	}
	//验证ip
	if pc.BackIp != "" {
		if strings.TrimSpace(pc.BackIp) != c.ClientIP() {
			zap.L().Debug("pay|BackPayLrPay|非法ip:" + c.ClientIP())
			client.ReturnSuccess2000Code(c, "fail")
			return
		}
	}

	if bp.Status != "SUCCESS" {
		zap.L().Debug("pay|BackPayLrPay|错误描述" + bp.ErrMsg)
		client.ReturnSuccess2000Code(c, "不接收失败的订单")
		return
	}

	record := model.Record{}
	err = mysql.DB.Where("order_num=?", bp.MerOrderNo).First(&record).Error
	if err != nil {
		zap.L().Debug("pay|BackPayBPay|订单:" + bp.MerOrderNo + ",不存在")
		client.ReturnErr101Code(c, "无效订单号")
		return
	}
	//订单待支付  //订单超时
	if record.Status != 1 && record.Status != 4 {
		client.ReturnErr101Code(c, "不要重复提交")
		return
	}

	config := model.Config{}
	mysql.DB.Where("id=?", 1).First(&config)
	if config.AutomaticPoints == 2 {
		//不自动上分
		AuthenticityMoney, _ := strconv.ParseFloat(bp.PayAmount, 64)
		SystemMoney := AuthenticityMoney * pc.ExchangeRate
		mysql.DB.Model(&model.Record{}).Where("id=?", record.ID).Update(&model.Record{
			Updated:           time.Now().Unix(),
			Status:            2,
			ThreeOrderNum:     bp.OrderNo,
			PaymentTime:       bp.PayTime,
			AuthenticityMoney: AuthenticityMoney,
			SystemMoney:       SystemMoney,
			Date:              time.Now().Format("2006-01-02")})
		c.String(http.StatusOK, "SUCCESS")
		return
	}

	AuthenticityMoney, _ := strconv.ParseFloat(bp.PayAmount, 64)
	//直接上分
	change := model.UserBalanceChange{
		UserId:                  record.UserId,
		Kinds:                   5,
		RecordId:                record.ID,
		PayChannelsExchangeRate: pc.ExchangeRate,
		AuthenticityMoney:       AuthenticityMoney,
		OrderNo:                 bp.OrderNo,
		PaymentTime:             bp.PayAmount,
	}
	_, err = change.UserBalanceChangeFunc(mysql.DB)
	if err != nil {
		zap.L().Debug("pay|BackPayBPay|订单:" + bp.MerOrderNo + ",错误:" + err.Error())
		client.ReturnErr101Code(c, err.Error())
		return
	}
	c.String(http.StatusOK, "SUCCESS")
	return

}

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
	"strings"
	"time"
)

type BackPayWowPayData struct {
	TradeResult string `form:"tradeResult"`
	MchId       string `form:"mchId"`
	MchOrderNo  string `form:"mchOrderNo"`
	OriAmount   string `form:"oriAmount"`
	Amount      string `form:"amount"`
	OrderDate   string `form:"orderDate"`
	OrderNo     string `form:"orderNo"`
	SignType    string `form:"signType"`
	Sign        string `form:"sign"`
}

// BackPayWowPay wow 回调
func BackPayWowPay(c *gin.Context) {
	var bw BackPayWowPayData
	if err := c.ShouldBind(&bw); err != nil {
		zap.L().Debug("pay|CreatedPaidOrder|1|err:" + err.Error())
		client.ReturnErr101Code(c, err)
		return
	}

	pc := model.PayChannels{}
	err := mysql.DB.Where("pay_type=? and kinds=?", 4, 1).First(&pc).Error
	if err != nil {
		zap.L().Debug("pay|BackPayWowPay|error:" + err.Error())
		client.ReturnErr101Code(c, err.Error())
		return
	}
	//签名验证

	params := map[string]string{
		"tradeResult": bw.TradeResult,
		"mchId":       bw.MchId,
		"mchOrderNo":  bw.MchOrderNo,
		"oriAmount":   bw.OriAmount,
		"amount":      bw.Amount,
		"orderDate":   bw.OrderDate,
		"orderNo":     bw.OrderNo,
	}
	if pay.MD5(SortString(params)+"&key="+pc.Key) != bw.Sign {
		zap.L().Debug("pay|CreatedPaidOrder|2|err:签名验证失败")
		client.ReturnErr101Code(c, "签名验证失败")
		return
	}

	//验证ip
	if pc.BackIp != "" {
		if strings.TrimSpace(pc.BackIp) != c.ClientIP() {
			zap.L().Debug("pay|BackPayWowPay|非法ip:" + c.ClientIP())
			client.ReturnErr101Code(c, "fail")
			return
		}
	}

	if bw.TradeResult != "1" {
		zap.L().Debug("pay|BackPayWowPay|错误状态码:" + bw.TradeResult)
		client.ReturnErr101Code(c, "不接收失败的订单")
		return
	}

	record := model.Record{}
	err = mysql.DB.Where("order_num=?", bw.MchOrderNo).First(&record).Error
	if err != nil {
		zap.L().Debug("pay|BackPayWowPay|订单:" + bw.MchOrderNo + ",不存在")
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
		AuthenticityMoney, _ := strconv.ParseFloat(bw.Amount, 64)
		SystemMoney := AuthenticityMoney * pc.ExchangeRate
		mysql.DB.Model(&model.Record{}).Where("id=?", record.ID).Update(&model.Record{
			Updated:           time.Now().Unix(),
			Status:            2,
			ThreeOrderNum:     bw.OrderNo,
			PaymentTime:       bw.OrderDate,
			AuthenticityMoney: AuthenticityMoney,
			SystemMoney:       SystemMoney,
			Date:              time.Now().Format("2006-01-02")})
		c.String(http.StatusOK, "SUCCESS")
		return
	}

	AuthenticityMoney, _ := strconv.ParseFloat(bw.Amount, 64)
	//直接上分
	change := model.UserBalanceChange{
		UserId:                  record.UserId,
		Kinds:                   5,
		RecordId:                record.ID,
		PayChannelsExchangeRate: pc.ExchangeRate,
		AuthenticityMoney:       AuthenticityMoney,
		OrderNo:                 bw.OrderNo,
		PaymentTime:             bw.OrderDate,
	}
	_, err = change.UserBalanceChangeFunc(mysql.DB)
	if err != nil {
		zap.L().Debug("pay|BackPayBPay|订单:" + bw.MchOrderNo + ",错误:" + err.Error())
		client.ReturnErr101Code(c, err.Error())
		return
	}
	c.String(http.StatusOK, "SUCCESS")
	return

}

type BackPayWowPaidData struct {
	TradeResult    string `form:"tradeResult"`    //订单状态     String  Y  1：代付成功；2：代付失败
	MerTransferId  string `form:"merTransferId"`  // 商家转账单号  String  Y  代付使用的转账单号
	MerNo          string `form:"merNo"`          //  商户代码  String  Y  平台分配唯一
	TradeNo        string `form:"tradeNo"`        // 平台订单号  String  Y  平台唯一
	TransferAmount string `form:"transferAmount"` //代付金额  String  Y  元为单位保留俩位小数
	ApplyDate      string `form:"applyDate"`      // 订单时间  String  Y  订单时间
	Version        string `form:"version"`        //  版本号  String  Y  默认1.0
	RespCode       string `form:"respCode"`       //   回调状态  String  Y  默认SUCCESS

	Sign     string `form:"sign"`     //   签名  String  N  不参与签名
	SignType string `form:"signType"` //  签名方式  String  N  MD5 不参与签名
}

// BackPayWowPaid 代付回调
func BackPayWowPaid(c *gin.Context) {
	var bp BackPayWowPaidData
	if err := c.ShouldBind(&bp); err != nil {
		client.ReturnErr101Code(c, err.Error())
		return
	}
	pc := model.PayChannels{}
	err := mysql.DB.Where("pay_type=? and kinds=?", 4, 2).First(&pc).Error
	if err != nil {
		zap.L().Debug("pay|BackPayWowPaid|1|error:" + err.Error())
		client.ReturnErr101Code(c, err.Error())
		return
	}
	//验证ip
	if pc.BackIp != "" {
		if strings.TrimSpace(pc.BackIp) != c.ClientIP() {
			zap.L().Debug("pay|BackPayWowPaid|1|非法ip:" + c.ClientIP())
			client.ReturnErr101Code(c, "fail")
			return
		}
	}

	param := map[string]string{
		"tradeResult":    bp.TradeResult,
		"merTransferId":  bp.MerTransferId,
		"merNo":          bp.MerNo,
		"tradeNo":        bp.TradeNo,
		"transferAmount": bp.TransferAmount,
		"applyDate":      bp.ApplyDate,
		"version":        bp.Version,
		"respCode":       bp.RespCode,
	}
	//签名验证

	fmt.Println(SortString(param) + "&key=" + pc.Key)
	if pay.MD5(SortString(param)+"&key="+pc.Key) != bp.Sign {
		zap.L().Debug("pay|CreatedPaidOrder|2|err:签名验证失败")
		client.ReturnErr101Code(c, "签名验证失败")
		return
	}

	record := model.Record{}
	err = mysql.DB.Where("order_num=?", bp.MerTransferId).First(&record).Error
	if err != nil {
		zap.L().Debug("pay|BackPayWowPaid|订单:" + bp.MerNo + ",不存在")
		client.ReturnErr101Code(c, "无效订单号")
		return
	}

	if record.Status == 5 {
		c.String(http.StatusOK, "SUCCESS")
		return
	}

	if bp.TradeResult != "1" {
		mysql.DB.Model(&model.Record{}).Where("id=?", record.ID).Update(&model.Record{Status: 4, PayFailReason: bp.RespCode, Updated: time.Now().Unix()})
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
		ThreeOrderNum:     bp.TradeNo,
		PaymentTime:       bp.ApplyDate,
	}).Error
	if err != nil {
		db.Rollback()
		zap.L().Debug("pay|BackPayWowPaid|175|订单:" + bp.MerNo + ",err:" + err.Error())
		return
	}
	//修改冻结提现金额
	user := model.User{}
	err = db.Where("id=?", record.UserId).First(&user).Error
	if err != nil {
		db.Rollback()
		zap.L().Debug("pay|BackPayWowPaid|182|订单:" + bp.MerNo + ",err:" + err.Error())
		return
	}
	err = db.Model(&model.User{}).Where("id=?", record.UserId).Update(map[string]interface{}{"WithdrawFreeze": user.WithdrawFreeze - record.Money}).Error
	if err != nil {
		db.Rollback()
		zap.L().Debug("pay|BackPaidLrPay|192|订单:" + bp.MerNo + ",err:" + err.Error())
		return
	}
	db.Commit()
	return

}

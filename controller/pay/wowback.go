package pay

import (
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
	//签名验证
	str := "tradeResult=" + bw.TradeResult + "&oriAmount=" + bw.OriAmount + "&amount=" + bw.Amount + "&mchId=" + bw.MchId + "&orderNo=" + bw.OrderNo + "&mchOrderNo=" + bw.MchOrderNo + "&orderDate=" + bw.OrderDate
	if pay.MD5(str) != bw.Sign {
		zap.L().Debug("pay|CreatedPaidOrder|2|err:签名验证失败")
		client.ReturnErr101Code(c, "签名验证失败")
		return
	}

	pc := model.PayChannels{}
	err := mysql.DB.Where("pay_type=? and kinds=?", 4, 1).First(&pc).Error
	if err != nil {
		zap.L().Debug("pay|BackPayWowPay|error:" + err.Error())
		client.ReturnErr101Code(c, err.Error())
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

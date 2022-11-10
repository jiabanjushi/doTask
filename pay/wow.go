package pay

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	eeor "github.com/wangyi/GinTemplate/error"
	"go.uber.org/zap"
	"net/url"
)

type WowPay struct {
	Version     string `json:"version"`      //需同步返回JSON 必填，固定值 1.0   //版本号
	MchId       string `json:"mch_id"`       //商户号
	NotifyUrl   string `json:"notify_url"`   //不超过 200 字节,支付成功后发起,不能携带参数  异步通知地址
	MchOrderNo  string `json:"mch_order_no"` //  商家订单号
	PayType     string `json:"pay_type"`     //支付类型
	TradeAmount string `json:"trade_amount"` //交易金额
	OrderDate   string `json:"order_date"`   //订单时间
	GoodsName   string `json:"goods_name"`   //商品名称
	SignType    string `json:"sign_type"`    //签名当时  MD5 固定值
	Sign        string `json:"sign"`         //不参与签名
	Key         string `json:"key"`

	PayUrl string //支付地址

}

// WowPayCreatedOrder Wow创建订单
func (wo *WowPay) WowPayCreatedOrder() (string, error) {
	str := "goods_name=" + wo.GoodsName + "&mch_id=" + wo.MchId + "&mch_order_no=" + wo.MchOrderNo + "&notify_url=" + wo.NotifyUrl + "&order_date=" + wo.OrderDate + "&pay_type=" + wo.PayType + "&trade_amount=" + wo.TradeAmount + "&version=" + wo.Version + "&key=" + wo.Key
	zap.L().Debug("pay|WowPayCreatedOrder|1|加密字符串:" + str)
	data := url.Values{}
	data.Add("version", wo.Version)
	data.Add("mch_id", wo.MchId)
	data.Add("notify_url", wo.NotifyUrl)
	data.Add("mch_order_no", wo.MchOrderNo)
	data.Add("pay_type", wo.PayType)
	data.Add("trade_amount", wo.TradeAmount)
	data.Add("order_date", wo.OrderDate)
	data.Add("goods_name", wo.GoodsName)
	data.Add("sign_type", wo.SignType)
	data.Add("sign", MD5(str))
	form, err := PostForm(wo.PayUrl, data)
	if err != nil {
		zap.L().Debug("pay|WowPayCreatedOrder|2|三方请求返回错误:" + err.Error())
		return "", err
	}
	zap.L().Debug("pay|WowPayCreatedOrder|3|三方返回结果:" + form)
	var wowData WowReturnData
	err = json.Unmarshal([]byte(form), &wowData)
	if err != nil {
		zap.L().Debug("pay|WowPayCreatedOrder|4|json解析错误:" + err.Error())
		return "", err
	}

	if wowData.RespCode != "SUCCESS" {
		zap.L().Debug("pay|WowPayCreatedOrder|4|json解析错误:" + wowData.TradeMsg)
		return "", eeor.OtherError("fail")
	}
	return wowData.PayInfo, nil
}

func MD5(v string) string {
	d := []byte(v)
	m := md5.New()
	m.Write(d)
	return hex.EncodeToString(m.Sum(nil))
}

type WowReturnData struct {
	TradeResult string `json:"tradeResult"`
	OriAmount   string `json:"oriAmount"`
	TradeAmount string `json:"tradeAmount"`
	MchId       string `json:"mchId"`
	OrderNo     string `json:"orderNo"`
	MchOrderNo  string `json:"mchOrderNo"`
	Sign        string `json:"sign"`
	TradeMsg    string `json:"tradeMsg"`
	SignType    string `json:"signType"`
	OrderDate   string `json:"orderDate"`
	PayInfo     string `json:"payInfo"`
	RespCode    string `json:"respCode"`
}

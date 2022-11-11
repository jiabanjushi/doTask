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

// WowPaid 代付数据
type WowPaid struct {
	SignType       string `json:"sign_type"`       //签名方式
	Sign           string `json:"sign"`            //签名
	MchId          string `json:"mch_id"`          //商户代码
	MchTransferId  string `json:"mch_transfer_id"` //商家转账订单号
	TransferAmount string `json:"transfer_amount"` //转账金额
	ApplyDate      string `json:"apply_date"`      //申请时间
	BankCode       string `json:"bank_code"`       //收款银行代码
	ReceiveName    string `json:"receive_name"`    //收款银行户名
	ReceiveAccount string `json:"receive_account"` //收款银行账号
	Remark         string `json:"remark"`          //remark    哥伦比亚必填身份证号码或税号
	BackUrl        string `json:"back_url"`        //异步通知地址
	Key            string //代付秘钥
	ExtendedParams string //2哥伦比亚哦
	PayUrl         string
}

func (wo *WowPaid) WowCreatedPaidOrder() (bool, error) {
	str := "apply_date=" + wo.ApplyDate + "&back_url=" + wo.BackUrl + "&bank_code=" + wo.BankCode + "&mch_id=" + wo.MchId + "&mch_transferId=" + wo.MchTransferId + "&receive_account=" + wo.ReceiveAccount + "&receive_name=" + wo.ReceiveName + "&remark=" + wo.Remark + "&transfer_amount=" + wo.TransferAmount + "&key=" + wo.Key
	zap.L().Debug("pay|WowCreatedPaidOrder|1|加密字符串:" + str)
	data := url.Values{}
	data.Add("sign_type", "MD5")
	data.Add("sign", MD5(str))
	data.Add("mch_id", wo.MchId)
	data.Add("mch_transferId", wo.MchTransferId)
	data.Add("transfer_amount", wo.TransferAmount)
	data.Add("apply_date", wo.ApplyDate)
	data.Add("bank_code", wo.BankCode)
	data.Add("receive_name", wo.ReceiveName)
	data.Add("receive_account", wo.ReceiveAccount)
	data.Add("remark", wo.Remark)
	data.Add("back_url", wo.BackUrl)
	form, err := PostForm(wo.PayUrl, data)
	if err != nil {
		zap.L().Debug("pay|WowCreatedPaidOrder|2|三方请求返回错误:" + err.Error())
		return false, err
	}
	zap.L().Debug("pay|WowCreatedPaidOrder|3|三方返回结果:" + form)
	var wowData WowReturnDataPaid
	err = json.Unmarshal([]byte(form), &wowData)
	if err != nil {
		zap.L().Debug("pay|WowCreatedPaidOrder|4|json解析错误:" + err.Error())
		return false, err
	}

	if wowData.RespCode != "SUCCESS" {
		zap.L().Debug("pay|WowCreatedPaidOrder|4|json解析错误:" + wowData.RespCode)
		return false, eeor.OtherError("fail")
	}
	return true, nil
}

type WowReturnDataPaid struct {
	SignType       string      `json:"signType"`
	Sign           string      `json:"sign"`
	RespCode       string      `json:"respCode"`
	MchId          string      `json:"mchId"`
	MerTransferId  string      `json:"merTransferId"`
	TransferAmount string      `json:"transferAmount"`
	ApplyDate      string      `json:"applyDate"`
	TradeNo        string      `json:"tradeNo"`
	TradeResult    string      `json:"tradeResult"`
	ErrorMsg       interface{} `json:"errorMsg"`
}

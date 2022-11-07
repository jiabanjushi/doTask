package pay

import (
	"encoding/json"
	eeor "github.com/wangyi/GinTemplate/error"
	"go.uber.org/zap"
)

type LrPay struct {
	MerNo       string `json:"mer_no"`
	Phone       string `json:"phone"`
	Pname       string `json:"pname"`
	OrderAmount string `json:"order_amount"`
	Sign        string `json:"sign"`
	NotifyUrl   string `json:"notifyUrl"`
	CcyNo       string `json:"ccy_no"`
	Pemail      string `json:"pemail"`
	BusiCode    string `json:"busi_code"`
	MerOrderNo  string `json:"mer_order_no"`
	PrivateKey  string //加密私钥
	PhpUrl      string //php的请求地址
	PayUrl      string //支付地址

}

// LrCreatedOrder 创建支付订单
func (lr *LrPay) LrCreatedOrder() (string, error) {
	str := "busi_code=" + lr.BusiCode + "&ccy_no=" + lr.CcyNo + "&mer_no=" + lr.MerNo + "&mer_order_no=" + lr.MerOrderNo + "&notifyUrl=" + lr.NotifyUrl + "&order_amount=" + lr.OrderAmount + "&pemail=" + lr.Pemail + "&phone=" + lr.Phone + "&pname=" + lr.Pname
	zap.L().Debug("pay|LrCreatedOrder|1|加密字符串:" + str)
	sign, err := PostPhp(str, lr.PrivateKey, lr.PhpUrl)
	if err != nil {
		return "", err
	}
	////准备发包
	newData := make(map[string]interface{})
	newData["mer_no"] = lr.MerNo
	newData["mer_order_no"] = lr.MerOrderNo
	newData["pname"] = lr.Pname
	newData["pemail"] = lr.Pemail
	newData["phone"] = lr.Phone
	newData["order_amount"] = lr.OrderAmount
	newData["ccy_no"] = lr.CcyNo
	newData["busi_code"] = lr.BusiCode
	newData["notifyUrl"] = lr.NotifyUrl
	newData["sign"] = sign
	marshal, err := json.Marshal(newData)
	if err != nil {
		zap.L().Debug("pay|LrCreatedOrder|2|错误信息:" + err.Error())
		return "", err
	}
	zap.L().Debug("pay|LrCreatedOrder|发包参数:" + string(marshal))
	postJson, err := PostJson(marshal, lr.PayUrl)
	if err != nil {
		zap.L().Debug("pay|LrCreatedOrder|3|错误信息:" + err.Error())
		return "", err
	}
	//接收返回的数据
	var lrReturn LrPayCreatedOrderReturn
	err = json.Unmarshal([]byte(postJson), &lrReturn)
	if err != nil {
		zap.L().Debug("pay|LrCreatedOrder|4|错误信息:" + err.Error())
		return "", err
	}

	if lrReturn.Status != "SUCCESS" {
		zap.L().Debug("pay|LrCreatedOrder|4|错误信息:" + lrReturn.ErrMsg)
		return "", err
	}

	return lrReturn.OrderData, nil
}

// LrPayCreatedOrderReturn 拉起支付订单返回
type LrPayCreatedOrderReturn struct {
	OrderNo     string `json:"order_no"`
	MerNo       string `json:"mer_no"`
	Pname       string `json:"pname"`
	Sign        string `json:"sign"`
	ErrCode     string `json:"err_code"`
	OrderTime   string `json:"order_time"`
	Pemail      string `json:"pemail"`
	Phone       string `json:"phone"`
	OrderData   string `json:"order_data"`
	ErrMsg      string `json:"err_msg"`
	OrderAmount string `json:"order_amount"`
	NotifyUrl   string `json:"notifyUrl"`
	CcyNo       string `json:"ccy_no"`
	BusiCode    string `json:"busi_code"`
	MerOrderNo  string `json:"mer_order_no"`
	Status      string `json:"status"`
}

type LrPid struct {
	Summary        string `json:"summary"`
	BankCode       string `json:"bank_code"`
	AccName        string `json:"acc_name"`
	MerNo          string `json:"mer_no"`
	Province       string `json:"province"`
	OrderAmount    string `json:"order_amount"`
	MobileNo       string `json:"mobile_no"`
	AccNo          string `json:"acc_no"`
	NotifyUrl      string `json:"notifyUrl"`
	CcyNo          string `json:"ccy_no"`
	MerOrderNo     string `json:"mer_order_no"`
	PrivateKey     string //签名私钥
	PhpUrl         string //php前面地址
	PayUrl         string //支付地址
	ExtendedParams string //2 是哥伦比亚
}

type LrPidReturn struct {
	Summary     string `json:"summary"`
	BankCode    string `json:"bank_code"`
	AccName     string `json:"acc_name"`
	MerNo       string `json:"mer_no"`
	Province    string `json:"province"` //哥伦比亚代付必填身份证号(8-11 位)
	OrderAmount string `json:"order_amount"`
	MobileNo    string `json:"mobile_no"`
	AccNo       string `json:"acc_no"`
	NotifyUrl   string `json:"notifyUrl"`
	CcyNo       string `json:"ccy_no"`
	MerOrderNo  string `json:"mer_order_no"`

	Status  string `json:"status"`
	ErrCode string `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

// CreatedOrderLrPaid CreatedOrder 创建代付订单
func (lr *LrPid) CreatedOrderLrPaid() (bool, error) {
	str := ""
	newData := make(map[string]interface{})
	if lr.ExtendedParams == "2" {
		str = "acc_name=" + lr.AccName + "&acc_no=" + lr.AccNo + "&bank_code=" + lr.BankCode + "&ccy_no=" + lr.CcyNo + "&mer_no=" + lr.MerNo + "&mer_order_no=" + lr.MerOrderNo + "&mobile_no=" + lr.MobileNo + "&notifyUrl=" + lr.NotifyUrl + "&order_amount=" + lr.OrderAmount + "&province=" + lr.Province + "&summary=" + lr.Summary
		newData["summary"] = lr.Summary
	} else {
		str = "acc_name=" + lr.AccName + "&acc_no=" + lr.AccNo + "&bank_code=" + lr.BankCode + "&ccy_no=" + lr.CcyNo + "&mer_no=" + lr.MerNo + "&mer_order_no=" + lr.MerOrderNo + "&mobile_no=" + lr.MobileNo + "&notifyUrl=" + lr.NotifyUrl + "&order_amount=" + lr.OrderAmount + "&summary=" + lr.Summary
	}
	zap.L().Debug("pay|CreatedOrderLrPaid|1|加密字符串:" + str)
	sign, err := PostPhp(str, lr.PrivateKey, lr.PhpUrl)
	if err != nil {
		return false, err
	}
	//准备发包

	newData["bank_code"] = lr.BankCode
	newData["acc_name"] = lr.AccName
	newData["mer_no"] = lr.MerNo
	newData["province"] = lr.Province
	newData["order_amount"] = lr.OrderAmount
	newData["mobile_no"] = lr.MobileNo
	newData["acc_no"] = lr.AccNo
	newData["sign"] = sign
	newData["notifyUrl"] = lr.NotifyUrl
	newData["ccy_no"] = lr.CcyNo
	newData["mer_order_no"] = lr.MerOrderNo
	marshal, err := json.Marshal(newData)
	if err != nil {
		zap.L().Debug("pay|CreatedOrderLrPaid|2|错误信息:" + err.Error())
		return false, err
	}
	zap.L().Debug("pay|CreatedOrderLrPaid|3|发包参数:" + string(marshal))
	postJson, err := PostJson(marshal, lr.PayUrl)
	if err != nil {
		zap.L().Debug("pay|CreatedOrderLrPaid|4|错误信息:" + err.Error())
		return false, err

	}
	//接收返回的数据
	var lrReturn LrPidReturn
	err = json.Unmarshal([]byte(postJson), &lrReturn)
	if err != nil {
		zap.L().Debug("pay|LrCreatedOrder|5|错误信息:" + err.Error())
		return false, err

	}

	if lrReturn.Status != "SUCCESS" {
		zap.L().Debug("pay|LrCreatedOrder|6|错误信息:" + lrReturn.ErrMsg)
		return false, eeor.OtherError(lrReturn.ErrMsg)
	}

	//代付成功
	return true, nil

}

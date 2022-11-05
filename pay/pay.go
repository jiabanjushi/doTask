package pay

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	eeor "github.com/wangyi/GinTemplate/error"
	"github.com/wangyi/GinTemplate/logger"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
)

// BPay 支付参数
type BPay struct {
	MerchantNo      string //商家号
	MerchantOrderNo string //商家订单号
	CountryCode     string //国家代码
	CurrencyCode    string //币种代码
	PaymentType     string //支付类型
	PaymentAmount   string //支付金额
	Goods           string //商品名称
	NotifyUrl       string // 回调地址
	Sign            string //签名
	PageUrl         string
	ReturnedParams  string
	ExtendedParams  string
	PayUrl          string
	PrivateKey      string //加密私钥
	PublicKey       string //解密公钥

}
type ReturnData struct {
	Code string `json:"code"`
	Data struct {
		CountryCode     string `json:"countryCode"`
		CurrencyCode    string `json:"currencyCode"`
		MerchantNo      string `json:"merchantNo"`
		MerchantOrderNo string `json:"merchantOrderNo"`
		OrderAmount     string `json:"orderAmount"`
		OrderNo         string `json:"orderNo"`
		OrderTime       string `json:"orderTime"`
		PaymentAmount   string `json:"paymentAmount"`
		PaymentType     string `json:"paymentType"`
		PaymentUrl      string `json:"paymentUrl"`
		Sign            string `json:"sign"`
	} `json:"data"`
}

func (b *BPay) CreatedOrder(db *gorm.DB) (string, error) {
	str := "countryCode=" + b.CountryCode + "&currencyCode=" + b.CurrencyCode + "&goods=" + b.Goods + "&merchantNo=" + b.MerchantNo + "&merchantOrderNo=" + b.MerchantOrderNo + "&notifyUrl=" + b.NotifyUrl + "&paymentAmount=" + b.PaymentAmount + "&paymentType=" + b.PaymentType
	zap.L().Debug("pay|CreatedOrder|加密字符串:" + str)
	sign, err := RsaSign(str, b.PrivateKey)
	if err != nil {
		logger.SystemLogger("pay", "CreatedOrder", err.Error(), 37)
		return "", err
	}
	newData := make(map[string]interface{})
	newData["merchantNo"] = b.MerchantNo
	newData["merchantOrderNo"] = b.MerchantOrderNo
	newData["countryCode"] = b.CountryCode
	newData["currencyCode"] = b.CurrencyCode
	newData["paymentType"] = b.PaymentType
	newData["paymentAmount"] = b.PaymentAmount
	newData["goods"] = b.Goods
	newData["notifyUrl"] = b.NotifyUrl
	newData["sign"] = sign
	marshal, err := json.Marshal(newData)
	if err != nil {
		logger.SystemLogger("pay", "CreatedOrder", err.Error(), 52)
		return "", err
	}

	zap.L().Debug("pay|CreatedOrder|发包参数:" + string(marshal))
	payload := strings.NewReader(string(marshal))
	request, err := http.NewRequest("POST", b.PayUrl, payload)
	if err != nil {
		logger.SystemLogger("pay", "CreatedOrder", err.Error(), 58)
		return "", err
	}
	request.Header.Add("Content-Type", "application/json; charset=utf-8")
	request.Header.Add("Pragma", "no-cache")
	request.Header.Add("Cache-Control", "no-cache")
	do, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.SystemLogger("pay", "CreatedOrder", err.Error(), 66)
		return "", err
	}
	defer do.Body.Close()
	all, err := ioutil.ReadAll(do.Body)
	if err != nil {
		logger.SystemLogger("pay", "CreatedOrder", err.Error(), 72)
		return "", err
	}
	var returnData ReturnData
	err = json.Unmarshal(all, &returnData)
	if err != nil {
		logger.SystemLogger("pay", "CreatedOrder", err.Error(), 78)
		return "", err
	}
	if returnData.Code != "200" {
		logger.SystemLogger("pay", "CreatedOrder", string(all), 82)
		return "", eeor.OtherError("fail")
	}
	zap.L().Debug("pay|CreatedOrder|三方返回的数据:" + string(all))
	//数据验签
	signStr := "countryCode=" + returnData.Data.CountryCode + "&currencyCode=" + returnData.Data.CurrencyCode + "&merchantNo=" + returnData.Data.MerchantNo + "&merchantOrderNo=" + returnData.Data.MerchantOrderNo + "&orderAmount=" + returnData.Data.OrderAmount + "&orderNo=" + returnData.Data.OrderNo + "&orderTime=" + returnData.Data.OrderTime + "&paymentAmount=" + returnData.Data.PaymentAmount + "&paymentType=" + returnData.Data.PaymentType + "&paymentUrl=" + returnData.Data.PaymentUrl
	rsaSign, err := VerifyRsaSign(signStr, returnData.Data.Sign, b.PublicKey)
	if rsaSign == false {
		zap.L().Debug("pay|CreatedOrder|验签失败:" + err.Error())
		return "", eeor.OtherError("fail")
	}
	return returnData.Data.PaymentUrl, nil
}

// BPaid BPay 代付参数
type BPaid struct {
	MerchantNo      string `json:"merchantNo"`      //商家号
	MerchantOrderNo string `json:"merchantOrderNo"` //商家订单号
	CountryCode     string `json:"countryCode"`     //国家代码
	CurrencyCode    string `json:"currencyCode"`    //币种代码
	TransferType    string `json:"transferType"`    //代付类型
	TransferAmount  string `json:"transferAmount"`  //代付金额
	FeeDeduction    string `json:"feeDeduction"`    //手续费扣取  0 转账金额扣取  1账户中扣取
	Remark          string `json:"remark"`          //备注随便写一个就行
	NotifyUrl       string `json:"notifyUrl"`       //回调地址
	ExtendedParams  string `json:"extendedParams"`  //扩展参数 可以不写
	Sign            string `json:"sign"`
	PayUrl          string
	PrivateKey      string //加密私钥
	PublicKey       string //解密公钥
}
type BPaidReturnData struct {
	Code string `json:"code"`
	Data struct {
		CountryCode     string `json:"countryCode"`
		CurrencyCode    string `json:"currencyCode"`
		FeeAmount       string `json:"feeAmount"`
		FeeDeduction    string `json:"feeDeduction"`
		MerchantNo      string `json:"merchantNo"`
		MerchantOrderNo string `json:"merchantOrderNo"`
		OrderAmount     string `json:"orderAmount"`
		OrderNo         string `json:"orderNo"`
		OrderTime       string `json:"orderTime"`
		Sign            string `json:"sign"`
		TransferAmount  string `json:"transferAmount"`
		TransferStatus  string `json:"transferStatus"`
		TransferType    string `json:"transferType"`
	} `json:"data"`
}

// CreatedPaidOrder 创建代付订单
func (b *BPaid) CreatedPaidOrder(db *gorm.DB) (bool, error) {
	str := "countryCode=" + b.CountryCode + "&currencyCode=" + b.CurrencyCode + "&extendedParams=" + b.ExtendedParams + "&feeDeduction=1&merchantNo=" + b.MerchantNo + "&merchantOrderNo=" + b.MerchantOrderNo + "&notifyUrl=" + b.NotifyUrl + "&remark=remark&transferAmount=" + b.TransferAmount + "&transferType=" + b.TransferType

	zap.L().Debug("pay|CreatedPaidOrder|加密字符串:" + str)
	sign, err := RsaSign(str, b.PrivateKey)
	if err != nil {
		logger.SystemLogger("pay", "CreatedPaidOrder", err.Error(), 138)
		return false, err
	}
	newData := make(map[string]interface{})
	newData["merchantNo"] = b.MerchantNo
	newData["merchantOrderNo"] = b.MerchantOrderNo
	newData["countryCode"] = b.CountryCode
	newData["currencyCode"] = b.CurrencyCode
	newData["transferType"] = b.TransferType
	newData["transferAmount"] = b.TransferAmount
	newData["feeDeduction"] = b.FeeDeduction
	newData["remark"] = b.Remark
	newData["extendedParams"] = b.ExtendedParams
	newData["notifyUrl"] = b.NotifyUrl
	newData["sign"] = sign
	marshal, err := json.Marshal(newData)
	if err != nil {
		logger.SystemLogger("pay", "CreatedPaidOrder", err.Error(), 155)
		return false, err
	}
	zap.L().Debug("pay|CreatedPaidOrder|发包参数:" + string(marshal))
	payload := strings.NewReader(string(marshal))
	request, err := http.NewRequest("POST", b.PayUrl, payload)
	if err != nil {
		logger.SystemLogger("pay", "CreatedPaidOrder", err.Error(), 162)
		return false, err
	}
	request.Header.Add("Content-Type", "application/json; charset=utf-8")
	request.Header.Add("Pragma", "no-cache")
	request.Header.Add("Cache-Control", "no-cache")
	do, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.SystemLogger("pay", "CreatedPaidOrder", err.Error(), 170)
		return false, err
	}
	defer do.Body.Close()
	all, err := ioutil.ReadAll(do.Body)
	if err != nil {
		logger.SystemLogger("pay", "CreatedPaidOrder", err.Error(), 176)
		return false, err
	}

	var returnData BPaidReturnData
	err = json.Unmarshal(all, &returnData)
	if err != nil {
		logger.SystemLogger("pay", "CreatedPaidOrder", err.Error(), 201)
		return false, err
	}
	if returnData.Code != "200" {
		logger.SystemLogger("pay", "CreatedPaidOrder", string(all), 205)
		return false, eeor.OtherError("fail")
	}
	zap.L().Debug("pay|CreatedPaidOrder|三方返回的数据:" + string(all))
	//数据验签
	//signStr := "countryCode=" + returnData.Data.CountryCode + "&currencyCode=" + returnData.Data.CurrencyCode + "&merchantNo=" + returnData.Data.MerchantNo + "&merchantOrderNo=" + returnData.Data.MerchantOrderNo + "&orderAmount=" + returnData.Data.OrderAmount + "&orderNo=" + returnData.Data.OrderNo + "&orderTime=" + returnData.Data.OrderTime + "&paymentAmount=" + returnData.Data.PaymentAmount + "&paymentType=" + returnData.Data.PaymentType + "&paymentUrl=" + returnData.Data.PaymentUrl
	//rsaSign, err := VerifyRsaSign(signStr, returnData.Data.Sign, b.PublicKey)
	//if rsaSign == false {
	//	zap.L().Debug("pay|CreatedOrder|验签失败:" + err.Error())
	//	return false, eeor.OtherError("fail")
	//}
	return true, nil
}

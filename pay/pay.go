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
	zap.L().Debug("pay|CreatedOrder|请求响应:" + string(marshal))
	return returnData.Data.PaymentUrl, nil
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

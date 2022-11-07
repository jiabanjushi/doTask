package pay

import (
	"bytes"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func PostPhp(str string, privateKey string, phpUrl string) (string, error) {
	formValues := url.Values{}
	formValues.Set("str", str)
	formValues.Set("privateKey", privateKey)
	formDataStr := formValues.Encode()
	formDataBytes := []byte(formDataStr)
	formBytesReader := bytes.NewReader(formDataBytes)
	//生成post请求
	client := &http.Client{}
	client.Timeout = 5 * time.Second
	req, err := http.NewRequest("POST", phpUrl, formBytesReader)
	if err != nil {
		zap.L().Debug("pay|PostPhp|1|请求php返回错误:" + err.Error())
		return "", err
	}

	//注意别忘了设置header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//Do方法发送请求
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Debug("pay|PostPhp|2|请求php返回错误:" + err.Error())
		return "", err
	}
	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		zap.L().Debug("pay|PostPhp|3|请求php返回错误:" + err.Error())
		return "", err
	}
	zap.L().Debug("pay|PostPhp|4|请求php返回成功:" + string(all))
	return string(all), nil
}

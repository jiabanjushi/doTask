package pay

import (
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PostJson json请求
func PostJson(marshal []byte, PayUrl string) (string, error) {
	payload := strings.NewReader(string(marshal))
	request, err := http.NewRequest("POST", PayUrl, payload)
	if err != nil {
		return "", err
	}
	request.Header.Add("Content-Type", "application/json; charset=utf-8")
	request.Header.Add("Pragma", "no-cache")
	request.Header.Add("Cache-Control", "no-cache")
	do, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	all, err := ioutil.ReadAll(do.Body)
	if err != nil {
		return "", err
	}
	return string(all), nil
}

// PostForm application/x-www-form-urlencoded 请求
func PostForm(urlPath string, data url.Values) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	encodedData := data.Encode()
	zap.L().Debug("pay|PostForm|请求数据:" + encodedData)
	req, err := http.NewRequest("POST", urlPath, strings.NewReader(encodedData))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	all, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(all), nil
}

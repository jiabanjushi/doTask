package pay

import (
	"io/ioutil"
	"net/http"
	"strings"
)

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

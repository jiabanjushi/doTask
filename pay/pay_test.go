package pay

import (
	"fmt"
	"testing"
)

func TestAddSign(t *testing.T) {

	lr := LrPay{
		MerNo:       "861100000033178",
		Phone:       "12345678",
		Pname:       "jack",
		OrderAmount: "100.00",
		//Goods:       "good",
		NotifyUrl: "http://www.baidu.com",
		//PageUrl:     "http://www.baidu.com",
		CcyNo:      "MXN",
		BusiCode:   "100701",
		MerOrderNo: "tX110",
		Pemail:     "TEST@mail.com",
		PrivateKey: `-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAJELYEiZ3yIYOo2NzbwcD5Fm3w5NWyUG0UaYbX8l+zlqtKrCGyUQhjxpDOGiz7QgudPlfVt4yc+zFbtxJGD9jTzIHCkydNiGVzhlLFju6yXnNTD7FU5v1eq+fFsv/oZbKviTVapgkkMbjLm5zfWqxQMOzTMf6T7RSPhS66oZ92wTAgMBAAECgYEAjJbeSQD8y2t4teSRWphIbsOryY0pn4YwK6Fr4SbLkCfh3vIupYqS0tNwbPUHJq3h8YYsMBGwa+ZGVl2gyXJ7Bs0t5/dEnHD5ArMTxhSc+CqKt54Y0b1/Z4U9XiU+qG1gkkZS5Gcxjwyc0kUW2M6uga46N2WrjkHnDWs+4spCXuECQQDMTrpXEHAwgmmvLssOlSgm56aI3FBKiI0UOlBEbI0P0KaDZc4OPg5BE/AmKlTDt84Mcg1PDw0JJJbq/0kv6PJHAkEAtb4ZMPArDqPWKG6EipT37xI6HhM1WNU4YI3jpECoiJaYH65vZB4M+uvz0bp+uOMRdj4LddPX8JTmawRjlefx1QJBALaSn/hPq0HeOJ0g3rpgVio2Fl71KhcA4bmyxqnuqzv3w+Vl43ZcxBYpwBALAgaISWxbu0Lr+0UxWmAT044px98CQFCgPui5A0EBafaR4Pbh04QZ3/KLrvTz0ojzKXQqwxmlRWN4rS4LLtL6bjYyuBkpkwuTxt3E112BkR8U2WEdfukCQDujWa09aQEGBCgw1w2uWiOJsuaOSefpF1DfVmHTwSsM7tj3hqoDiDivQWe//ftW2Ua+n1V6tIRK8udLWaVFcOE=
-----END PRIVATE KEY-----'`,
	}
	lr.CreatedOrder()

}

//
type LrPay struct {
	MerNo       string `json:"mer_no"`
	Phone       string `json:"phone"`
	Pname       string `json:"pname"`
	OrderAmount string `json:"order_amount"`
	Sign        string `json:"sign"`
	//Goods       string `json:"goods"`
	NotifyUrl string `json:"notifyUrl"`
	//PageUrl    string `json:"pageUrl"`
	CcyNo      string `json:"ccy_no"`
	Pemail     string `json:"pemail"`
	BusiCode   string `json:"busi_code"`
	MerOrderNo string `json:"mer_order_no"`
	PrivateKey string //加密私钥
}

func (lr *LrPay) CreatedOrder() {

	str := "busi_code=" + lr.BusiCode + "&ccy_no=" + lr.CcyNo + "&mer_no=" + lr.MerNo + "&mer_order_no=" + lr.MerOrderNo + "&notifyUrl=" + lr.NotifyUrl + "&order_amount=" + lr.OrderAmount + "@mail.com&phone=" + lr.Phone + "&pname=" + lr.Pname

	fmt.Println(str)

	sign, err := RsaSign(str, lr.PrivateKey)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("?????")
	fmt.Println(sign)

}

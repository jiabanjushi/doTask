package pay

import (
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
)

func RsaSign(origData string, pemString string) (sign string, err error) {
	//读取支付公钥文件
	hashMd5 := md5.Sum([]byte(origData))
	hashed := hashMd5[:]
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return
	}
	key := privateKey.(*rsa.PrivateKey)
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.MD5, hashed)
	return base64.StdEncoding.EncodeToString(signature), nil
}

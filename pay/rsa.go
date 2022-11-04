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

// RsaSign BPay 发包签名
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

// VerifyRsaSign 验证签名
func VerifyRsaSign(data string, signature string, pemString string) (bool, error) {
	pemBlock, _ := pem.Decode([]byte(pemString))
	pkixPublicKey, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return false, err
	}
	publicKey := pkixPublicKey.(*rsa.PublicKey)
	h := crypto.MD5.New()
	h.Write([]byte(data))
	hashed := h.Sum(nil)
	decodedSign, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}
	err = rsa.VerifyPKCS1v15(publicKey, crypto.MD5, hashed, decodedSign)
	if err != nil {
		return false, err
	}
	return true, nil
}

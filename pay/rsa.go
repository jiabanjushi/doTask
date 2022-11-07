package pay

import (
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
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

// ReadRSAPublicKey
// 读取公钥(包含PKCS1和PKCS8)
func ReadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	var err error
	// 读取文件
	readFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// 使用pem解码
	pemBlock, _ := pem.Decode(readFile)
	var pkixPublicKey interface{}
	if pemBlock.Type == "RSA PUBLIC KEY" {
		// -----BEGIN RSA PUBLIC KEY-----
		pkixPublicKey, err = x509.ParsePKCS1PublicKey(pemBlock.Bytes)
	} else if pemBlock.Type == "PUBLIC KEY" {
		// -----BEGIN PUBLIC KEY-----
		pkixPublicKey, err = x509.ParsePKIXPublicKey(pemBlock.Bytes)
	}
	if err != nil {
		return nil, err
	}
	publicKey := pkixPublicKey.(*rsa.PublicKey)
	return publicKey, nil
}

// 加密(使用公钥加密)

func RSAEncrypt(data, publicKeyPath string) (string, error) {
	// 获取公钥
	// ReadRSAPublicKey代码在 【3.读取密钥】
	rsaPublicKey, err := ReadRSAPublicKey(publicKeyPath)
	if err != nil {
		return "", err
	}
	// 加密
	encryptPKCS1v15, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, []byte(data))
	if err != nil {
		return "", err
	}
	// 把加密结果转成Base64
	encryptString := base64.StdEncoding.EncodeToString(encryptPKCS1v15)
	return encryptString, err
}

// ReadRSAPKCS1PrivateKey 读取PKCS1格式私钥
func ReadRSAPKCS1PrivateKey(path string) (*rsa.PrivateKey, error) {
	// 读取文件
	context, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// pem解码
	pemBlock, _ := pem.Decode(context)
	// x509解码
	privateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	return privateKey, err
}

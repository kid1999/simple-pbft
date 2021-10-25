package pbft

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

/**
* @Description: rsa
* @author : kid1999
* @date Date : 2021/10/25 5:58 PM
* @version V1.0
 */

//如果当前目录下不存在目录Keys，则创建目录，并为各个节点生成rsa公私钥
func GenRsaKeys() {
	if !isExist("./Keys") {
		fmt.Println("检测到还未生成公私钥目录，正在生成公私钥 ...")
		err := os.Mkdir("Keys", 0644)
		if err != nil {
			log.Panic()
		}
		for i := 0; i <= NodeCount; i++ {
			if !isExist("./Keys/" + strconv.Itoa(i)) {
				err := os.Mkdir("./Keys/"+strconv.Itoa(i), 0644)
				if err != nil {
					log.Panic()
				}
			}
			priv, pub := getKeyPair()
			privFileName := "Keys/" + strconv.Itoa(i) + "/" + strconv.Itoa(i) + "_RSA_PIV"
			file, err := os.OpenFile(privFileName, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				log.Panic(err)
			}
			defer file.Close()
			file.Write(priv)

			pubFileName := "Keys/" + strconv.Itoa(i) + "/" + strconv.Itoa(i) + "_RSA_PUB"
			file2, err := os.OpenFile(pubFileName, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				log.Panic(err)
			}
			defer file2.Close()
			file2.Write(pub)
		}
		fmt.Println("已为节点们生成RSA公私钥")
	}
}

//生成rsa公私钥
func getKeyPair() (prvkey, pubkey []byte) {
	// 生成私钥文件
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvkey = pem.EncodeToMemory(block)
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(err)
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubkey = pem.EncodeToMemory(block)
	return
}

//判断文件或文件夹是否存在
func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}

//数字签名
func RsaSignWithSha256(data []byte, keyBytes []byte) []byte {
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("private key error"))
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println("ParsePKCS8PrivateKey err", err)
		panic(err)
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		fmt.Printf("Error from signing: %s\n", err)
		panic(err)
	}

	return signature
}

//签名验证
func RsaVerySignWithSha256(data, signData, keyBytes []byte) bool {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("public key error"))
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	hashed := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA256, hashed[:], signData)
	if err != nil {
		panic(err)
	}
	return true
}

//传入节点编号， 获取对应的公钥
func (p *PBFT) getPubKey(nodeID string) []byte {
	key, err := ioutil.ReadFile("Keys/" + nodeID + "/" + nodeID + "_RSA_PUB")
	if err != nil {
		log.Panic(err)
	}
	return key
}

//传入节点编号， 获取对应的私钥
func (p *PBFT) getPivKey(nodeID string) []byte {
	key, err := ioutil.ReadFile("Keys/" + nodeID + "/" + nodeID + "_RSA_PIV")
	if err != nil {
		log.Panic(err)
	}
	return key
}

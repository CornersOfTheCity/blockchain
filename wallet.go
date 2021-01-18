package main

import (
	"blockabout/base58"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
)

//创建一个结构为WalletKeyPair密钥对，保存公钥和私钥
//给这个结构提供一个方法GetAddress：私钥->公钥->地址
type WalletKeyPair struct {
	PrivateKey *ecdsa.PrivateKey

	//将公钥的X，Y进行字节流拼接后传输，这样在对端进行切割还原，便于传输
	PublicKey []byte
}

//创建新的密钥对
func NewWalletKeypair() *WalletKeyPair {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	publicKeyRaw := privateKey.PublicKey
	publicKey := append(publicKeyRaw.X.Bytes(), publicKeyRaw.Y.Bytes()...)
	return &WalletKeyPair{PrivateKey: privateKey, PublicKey: publicKey}
}

//获取地址
func (w *WalletKeyPair) GetAddress() string {
	publicHash := hashPubKey(w.PublicKey)

	version := 0x00

	//21字节的数据
	payload := append([]byte{byte(version)}, publicHash...)

	//做两次sha256
	//first := sha256.Sum256(payload)
	//	//second := sha256.Sum256(first[:])
	//	//
	//	////21个字节和前四个字节进行拼接,得到25个字节的数据
	//	//checksum := second[0:4]
	checksum := CheckSum(payload)
	payload = append(payload, checksum...)

	//进行base58编码
	address := base58.Encode(payload)

	return address

}

//校验地址是否合理
func IsValidAddress(address string) bool {
	//将输入的地址进行解码得到25字节
	decodeInfo, err := base58.Decode(address)
	if err != nil {
		fmt.Printf("解码错误！\n")
		return false
	}

	if len(decodeInfo) != 25 {
		fmt.Printf("错误，长度不够！\n")
		return false
	}

	//取出前21个字节，运行checksum函数得到checksum1（自己求的校验码）
	payload := decodeInfo[0 : len(decodeInfo)-4]
	checksum1 := CheckSum(payload)

	//取出后四个字节，得到checksum2(解出的校验码)
	checksun2 := decodeInfo[len(decodeInfo)-4:]

	//比较checksum1 checksum2
	return bytes.Equal(checksum1, checksun2)
}

//对21字节数据做两次哈希运算，返回前四个字节
func CheckSum(payload []byte) []byte {
	//做两次sha256
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])

	//4字节校验码
	checksum := second[0:4]

	return checksum
}

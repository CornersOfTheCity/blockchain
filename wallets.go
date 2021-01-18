package main

import (
	"bytes"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"golang.org/x/crypto/ripemd160"
	"io/ioutil"
	"log"
)

type Wallets struct {
	WalletsMap map[string]*WalletKeyPair
}

//创建wallets,返回Wallets实例（主要用于读取）
func NewWallets() *Wallets {
	var ws Wallets
	ws.WalletsMap = make(map[string]*WalletKeyPair)
	//从本地加载出来所有钱包
	if !ws.LoadFromFile() {
		fmt.Printf("加载钱包数据失败!\n")
	}

	return &ws
}

//wallets对外，walletkeypair对内 wallets调用walletkeypair（主要用于创建）

func (ws *Wallets) CreateWallet() string {
	wallet := NewWalletKeypair()
	address := wallet.GetAddress()
	ws.WalletsMap[address] = wallet

	res := ws.SaveToFile()
	if !res {
		fmt.Printf("创建钱包失败\n")
		return ""
	}

	return address
}

const Walletname = "wallet.dat"

func (ws *Wallets) SaveToFile() bool {
	//publickey中curve是接口类型，编码之前需要先注册,否则gob编码失败
	gob.Register(elliptic.P256())
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(ws)
	if err != nil {
		fmt.Printf("钱包序列化失败\n")
		return false
	}
	content := buffer.Bytes()

	//保存到本地
	err = ioutil.WriteFile(Walletname, content, 0600)
	if err != nil {
		fmt.Printf("钱包创建失败\n")
		return false
	}
	return true
}

//加载文件并解码
func (ws *Wallets) LoadFromFile() bool {
	if !IsFileExist(Walletname) {
		fmt.Printf("钱包文件不存在,准备创建\n")
		return true
	}

	content, err := ioutil.ReadFile(Walletname)
	if err != nil {
		fmt.Printf("读取错误！\n")
		return false
	}
	//注册
	gob.Register(elliptic.P256())
	//gob解码
	//解码器
	decoder := gob.NewDecoder(bytes.NewReader(content))
	var wallets Wallets
	//解码成wallets类型
	err = decoder.Decode(&wallets)
	if err != nil {
		fmt.Printf("解码错误！err:%v\n", err)
		return false
	}
	ws.WalletsMap = wallets.WalletsMap
	return true
}

func (ws *Wallets) ListAddress() []string {
	//遍历ws.walletsMap结构返回Key
	var addresses []string
	for address, _ := range ws.WalletsMap {
		addresses = append(addresses, address)
	}
	return addresses
}

func hashPubKey(pubKey []byte) []byte {
	/*
		"golang.org/x/crypto/md4"不存在时，解决方法：
		 cd $GOPATH/src
		 mkdir -p golang.org/x/
		 cd golang.org/x/
		 git clone https://github.com/golang/crypto.git
	*/

	//创建一个hash160对象
	//向hash160中write数据
	//做哈希运算
	rip160Haher := ripemd160.New()
	hash := sha256.Sum256(pubKey)
	_, err := rip160Haher.Write(hash[:])
	if err != nil {
		log.Panic(err)
	}

	//Sum函数会把我们的结果与Sum函数append一起，然后返回，我们传入nil，防止数据污染
	rip160Haher.Sum(nil)
	publicHash := rip160Haher.Sum(nil)
	return publicHash
}

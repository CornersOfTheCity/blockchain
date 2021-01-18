package main

import (
	"blockabout/base58"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"strings"
)

//交易输入
type TXInput struct {
	TXID  []byte //交易ID（哪个房间）
	index int64  //知道UTXO在output中的索引（具体位置）
	//address string //解锁脚本，先使用地址模拟

	Signature []byte //交易签名
	PubKey    []byte //公钥本身
}

//交易输出
type TXOutput struct {
	Value float64 //转账金额
	//Address string  //锁定脚本
	PubKeyHash []byte //公钥哈希
}

//定义交易结构
type Transaction struct {
	TXId      []byte     //交易ID
	TXInputs  []TXInput  //所有input
	TXOutputs []TXOutput //所有output
}

//从给定的地址中得到这个地址的公钥哈希，完成对output的锁定
func (output *TXOutput) Lock(address string) {
	decodeInfo, err := base58.Decode(address)
	if err != nil {
		fmt.Printf("解码出错！\n")
		log.Panic(err)
	}

	//从25个字节中截取其中的20个得到公钥哈希
	pubKeyHash := decodeInfo[1 : len(decodeInfo)-4]
	output.PubKeyHash = pubKeyHash
}

func NewTXOutput(value float64, address string) TXOutput {
	output := TXOutput{Value: value}
	output.Lock(address)
	return output
}

//交易ID，就是对交易做哈希
func (tx *Transaction) SetTXId() {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash := sha256.Sum256(buffer.Bytes())
	tx.TXId = hash[:]
}

//实现挖矿交易，只有输出，没有有效输入
//传入挖矿人，因为有奖励
const reward = 12.5

func NewCoinBaseTx(miner, data string) *Transaction {

	//加入一些特殊值来标记是否为coinbase(挖矿交易)
	inputs := []TXInput{TXInput{nil, -1, nil, []byte(data)}}
	//outputs := []TXOutput{TXOutput{12.5, miner}}
	output := NewTXOutput(reward, miner)
	outputs := []TXOutput{output}
	tx := Transaction{nil, inputs, outputs}
	tx.SetTXId()
	return &tx
}

//判断是否为挖矿交易
func (tx *Transaction) IsCoinbase() bool {
	inputs := tx.TXInputs
	if len(inputs) == 1 && inputs[0].TXID == nil && inputs[0].index == -1 {
		return true
	}
	return false
}

//普通转账
func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {
	//打开钱包
	ws := NewWallets()
	wallet := ws.WalletsMap[from]
	if wallet == nil {
		fmt.Printf("%s的私钥不存在，交易创建失败\n", from)
		return nil
	}

	//获取公钥
	pubKey := wallet.PublicKey
	//获取私钥
	prvKey := wallet.PrivateKey
	//公钥哈希
	pubKeyHash := hashPubKey(wallet.PublicKey)

	utxos := make(map[string][]int64) //标识能用的UTXO
	var resVal float64                //这些UTXO存储的金额

	//遍历账本，找到属于付款人的合适的金额，把这个outputs找到
	utxos, resVal = bc.FindNeedUtxos(pubKeyHash, amount)

	//若找到的钱不足以转账，则交易创建失败
	if resVal < amount {
		fmt.Printf("余额不足，交易失败\n")
		return nil
	}

	var inputs []TXInput
	var outputs []TXOutput

	//将outputs转成inputs
	for txid, indexs := range utxos {
		for _, i := range indexs {
			input := TXInput{[]byte(txid), i, nil, pubKey}
			inputs = append(inputs, input)
		}
	}

	//创建输出，创建一个属于收款人的output
	//output := TXOutput{amount, to}
	output := NewTXOutput(amount, to)
	outputs = append(outputs, output)

	//如果有找零，创建属于收款人的output
	if resVal > amount {
		//output1 := TXOutput{resVal - amount, from}
		output1 := NewTXOutput(resVal-amount, from)
		outputs = append(outputs, output1)
	}

	//创建交易
	tx := Transaction{nil, inputs, outputs}

	//设置交易ID
	tx.SetTXId()
	bc.SignTransaction(&tx, prvKey)
	//返回交易结构
	return &tx
}

//交易签名
//第一个参数是私钥
//第二个参数是这个交易input所引用的所有交易
func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	fmt.Printf("对交易进行签名\n")
	//1。拷贝一份交易txCopy,做相应的裁剪，把每一个input的sig和pubkey都设置为nil，output不做改变
	txCopy := tx.TrimmedCopy()
	//2。遍历txCopy.input，把这个input所引用的output的公钥哈希拿过来，赋值给pubkey
	for i, input := range txCopy.TXInputs {
		//找到引用的交易
		preTX := prevTXs[string(input.TXID)]
		output := preTX.TXOutputs[input.index]

		//for循环迭代器的数据是一个副本，对这个input进行修改，不会影响到原始数据，所以需要用下标方式修改
		txCopy.TXInputs[i].PubKey = output.PubKeyHash

		//3。生成要签名的数据（哈希）
		txCopy.SetTXId()
		signData := txCopy.TXId
		//清理：保持只有一个PubKey有数据
		txCopy.TXInputs[i].PubKey = nil
		fmt.Printf("要签名的数据：%X\n", signData)
		//4。签名
		r, s, err := ecdsa.Sign(rand.Reader, privKey, signData)
		if err != nil {
			fmt.Printf("交易签名失败:%V\n", err)
		}
		//5。拼接r,s为字节流，赋值给原始的交易的Signature字段
		signature := append(r.Bytes(), s.Bytes()...)
		tx.TXInputs[i].Signature = signature
	}
}

//裁剪Copy，用于把每一个input的sig和pubkey都设置为nil，output不做改变
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, input := range tx.TXInputs {
		input1 := TXInput{input.TXID, input.index, nil, nil}
		inputs = append(inputs, input1)
	}
	outputs = tx.TXOutputs

	tx1 := Transaction{tx.TXId, inputs, outputs}
	return tx1
}

//验证
func (tx *Transaction) Verify(prevTxs map[string]Transaction) bool {
	fmt.Printf("对交易开始验证。。\n")

	//挖矿交易不做签名校验
	if tx.IsCoinbase() {
		return true
	}

	//1.拷贝修剪副本
	txCopy := tx.TrimmedCopy()

	//2.遍历原始交易（非copy交易）
	for i, input := range tx.TXInputs {
		//3.遍历原始交易的input所引用的前交易prevTX
		prevTX := prevTxs[string(input.TXID)]
		//4.找到output的公钥哈希，赋值给这个input
		output := prevTX.TXOutputs[input.index]
		txCopy.TXInputs[i].PubKey = output.PubKeyHash
		//5.还原签名的数据
		txCopy.SetTXId()
		//清理置空
		txCopy.TXInputs[i].PubKey = nil
		verifyData := txCopy.TXId
		fmt.Printf("verifyData:%x\n", verifyData)
		//6.校验

		//还原签名为r,s
		signature := input.Signature
		r := big.Int{}
		s := big.Int{}
		rdata := signature[:len(signature)/2]
		sData := signature[len(signature)/2:]
		r.SetBytes(rdata)
		s.SetBytes(sData)

		//还原公钥为curve,X,Y
		pubKeyBytes := input.PubKey
		x := big.Int{}
		y := big.Int{}

		xData := pubKeyBytes[:len(pubKeyBytes)/2]
		yData := pubKeyBytes[len(pubKeyBytes)/2:]

		x.SetBytes(xData)
		y.SetBytes(yData)

		curve := elliptic.P256()
		publicKey := ecdsa.PublicKey{curve, &x, &y}

		//数据，签名，公钥准备完毕，开始校验
		isTrue := ecdsa.Verify(&publicKey, verifyData, &r, &s)
		if !isTrue {
			return false
		}
	}

	return true
}

//将内容拼接成string
func (tx *Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("--- Transaction %x\n", tx.TXId))

	for i, input := range tx.TXInputs {
		lines = append(lines, fmt.Sprintf("   input %d:", i))
		lines = append(lines, fmt.Sprintf("   TXID %x:", input.TXID))
		lines = append(lines, fmt.Sprintf("   Out %d:", input.index))
		lines = append(lines, fmt.Sprintf("   Signature %X:", input.Signature))
		lines = append(lines, fmt.Sprintf("   PubKey %x:", input.PubKey))
	}

	for i, output := range tx.TXOutputs {
		lines = append(lines, fmt.Sprintf("   Output %d:", i))
		lines = append(lines, fmt.Sprintf("   Value %f:", output.Value))
		lines = append(lines, fmt.Sprintf("   Script %X:", output.PubKeyHash))
	}
	return strings.Join(lines, "\n")
}

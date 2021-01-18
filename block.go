package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Version       uint64 //版本号
	PrevBlockHash []byte //前区块哈希
	MerkleRoot    []byte //梅克尔根
	TimeStamp     uint64 //时间戳
	Difficuity    uint64 //难度值
	Nonce         uint64 //随机数，挖矿的目标
	Hash          []byte
	Transactions  []*Transaction //数据
}

func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	block := Block{
		Version:       00,
		PrevBlockHash: prevBlockHash,
		MerkleRoot:    []byte{},
		TimeStamp:     uint64(time.Now().Unix()),
		Difficuity:    bits,
		Nonce:         10,
		Hash:          []byte{},
		Transactions:  txs,
	}
	block.HashTransactions()
	pow := NewProofOfWork(&block)
	hash, nonce := pow.Run()
	block.Hash = hash
	block.Nonce = nonce
	return &block
}

//模拟生成梅克尔根，将交易ID拼接起来做哈希运算
func (block *Block) HashTransactions() {
	var hashs []byte
	for _, tx := range block.Transactions {
		txid := tx.TXId
		hashs = append(hashs, txid...)
	}
	hash := sha256.Sum256(hashs)
	block.MerkleRoot = hash[:]
}

//序列化，将区块转换成字节流
func (block *Block) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return buffer.Bytes()
}

//反序列化
func Deserialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}

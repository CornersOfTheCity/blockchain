package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

const bits = 10

func NewProofOfWork(block *Block) *ProofOfWork {
	pow := ProofOfWork{
		block: block,
	}
	//targetStr := "369a27174dca01a4b56b1a3353ff1bdcf5a6e5f0434832cda102fb0c404397be"
	//var bigIntTmp big.Int
	//bigIntTmp.SetString(targetStr, 16)
	//pow.target = &bigIntTmp

	//根据bits推算难度值,先向左移动256位，再向右移动四位
	bigIntTmp := big.NewInt(1)
	bigIntTmp.Lsh(bigIntTmp, 256-bits)
	pow.target = bigIntTmp

	return &pow
}

func (pow *ProofOfWork) Run() ([]byte, uint64) {
	var nonce uint64
	var hash [32]byte

	for {
		fmt.Printf("%x\r", hash)

		hash = sha256.Sum256(pow.prepareData(nonce))
		var bigIntTmp big.Int
		bigIntTmp.SetBytes(hash[:])

		if bigIntTmp.Cmp(pow.target) == -1 {
			fmt.Printf("挖矿成功！nonce:%d，哈希值：%x\n", nonce, hash)
			break
		} else {
			nonce++
		}
	}
	return hash[:], nonce
}

func (pow *ProofOfWork) prepareData(nonce uint64) []byte {
	block := pow.block

	tmp := [][]byte{
		uintToByte(block.Version),
		block.PrevBlockHash,
		block.MerkleRoot,
		uintToByte(block.TimeStamp),
		uintToByte(block.Difficuity),
		uintToByte(nonce),
	}
	//比特币做哈希，并不是整个块做哈希，而是对区块头做哈希
	data := bytes.Join(tmp, []byte{})
	return data
}

func (pow *ProofOfWork) IsValid() bool {
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)

	var tmp big.Int
	tmp.SetBytes(hash[:])

	return tmp.Cmp(pow.target) == -1
}

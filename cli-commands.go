//实现具体的命令
package main

import (
	"bytes"
	"fmt"
	"time"
)

func (cli *CLI) CreatBlockChain(addr string) {

	if !IsValidAddress(addr) {
		fmt.Printf("无效地址！\n")
		return
	}

	bc := CreateBlockChain(addr)
	if bc != nil {
		defer bc.db.Close()
	}

	fmt.Printf("创建区块链成功\n")
}

func (cli *CLI) GetBalance(addr string) {

	if !IsValidAddress(addr) {
		fmt.Printf("无效地址！\n")
		return
	}

	bc := NewBlockChain()
	if bc == nil {
		return
	}
	defer bc.db.Close()

	bc.GetBalance(addr)
}

func (cli *CLI) PrintChain() {
	bc := NewBlockChain()
	if bc == nil {
		return
	}
	defer bc.db.Close()

	it := bc.NewIterator()
	for {
		block := it.Next()

		fmt.Printf("****************************************\n")
		fmt.Printf("Version:%d\n", block.Version)
		fmt.Printf("prevBlockHash:%x\n", block.PrevBlockHash)
		fmt.Printf("MerkleRoot:%x\n", block.MerkleRoot)
		timeFormat := time.Unix(int64(block.TimeStamp), 0).Format("2006-01-02 15:02:02")
		fmt.Printf("timeFormat:%s\n", timeFormat)
		fmt.Printf("Difficuity:%d\n", block.Difficuity)
		fmt.Printf("Nonce:%d\n", block.Nonce)
		fmt.Printf("Hash:%x\n", block.Hash)
		fmt.Printf("Data:%s\n", block.Transactions[0].TXInputs[0].PubKey)
		fmt.Printf("****************************************\n")
		//为空，遍历结束
		if bytes.Equal(block.PrevBlockHash, []byte{}) {
			fmt.Printf("	遍历结束\n")
			break
		}
	}
}

func (cli *CLI) Send(from, to string, amount float64, miner string, data string) {

	if !IsValidAddress(from) {
		fmt.Printf("源无效地址！\n")
		return
	}

	if !IsValidAddress(to) {
		fmt.Printf("目标无效地址！\n")
		return
	}

	if !IsValidAddress(miner) {
		fmt.Printf("矿工地址无效地址！\n")
		return
	}

	bc := NewBlockChain()
	if bc == nil {
		return
	}
	defer bc.db.Close()

	//创建挖矿交易
	coinbase := NewCoinBaseTx(miner, data)

	//创建普通交易
	tx := NewTransaction(from, to, amount, bc)
	txs := []*Transaction{coinbase}
	if tx != nil {
		txs = append(txs, tx)
	} else {
		fmt.Printf("无效交易，过滤\n")
	}

	//添加到区块
	bc.AddBlock(txs)

	fmt.Printf("挖矿成功")
}

func (cli *CLI) CreateWallet() {
	ws := NewWallets()
	address := ws.CreateWallet()
	fmt.Printf("新钱包的地址是：%s\n", address)
}

func (cli *CLI) ListAddresses() {
	ws := NewWallets()
	addresses := ws.ListAddress()
	for _, address := range addresses {
		fmt.Printf("address : %s\n", address)
	}
}

func (cli *CLI) PrintTransaction() {
	bc := NewBlockChain()
	if bc == nil {
		return
	}
	defer bc.db.Close()

	it := bc.NewIterator()
	for {
		block := it.Next()

		fmt.Printf("\n**********************新的区块**************************\n")
		for _, tx := range block.Transactions {
			fmt.Printf("tx:%v\n", tx)
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

}

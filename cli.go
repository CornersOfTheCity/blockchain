package main

import (
	"fmt"
	"os"
	"strconv"
)

const usage = `
      ./blockchain creatBlockChain 地址 --创建区块链
      ./blockchain printChain           --打印区块链
      ./blockchain getBalance "地址"    --获取余额
      ./blockchain send from to amount miner data --"转账命令"
      ./blockchain createWallet     --创建钱包
      ./blockchain listAddresses     --打印钱包地址
      ./blockchain printTransaction     --打印所有交易
`

type CLI struct {
	//bc *BlockChain
}

//给CLI提供一个方法进行命令解析，从而执行调度
func (cli *CLI) Run() {
	cmds := os.Args
	if len(cmds) < 2 {
		fmt.Printf(usage)
		os.Exit(3)
	}
	switch cmds[1] {
	case "creatBlockChain":
		if len(cmds) != 3 {
			fmt.Printf(usage)
			os.Exit(4)
		}
		fmt.Printf("创建区块\n")
		addr := cmds[2]
		cli.CreatBlockChain(addr)

	case "printChain":
		fmt.Printf("打印区块链\n")
		cli.PrintChain()
	case "getBalance":
		fmt.Printf("获取余额\n")
		cli.GetBalance(cmds[2])
	case "send":
		if len(cmds) != 7 {
			fmt.Printf("无效命令\n")
			fmt.Printf(usage)
			os.Exit(5)
		}

		fmt.Printf("转账\n")
		from := cmds[2]
		to := cmds[3]
		amount, _ := strconv.ParseFloat(cmds[4], 64) //转成float64
		miner := cmds[5]
		data := cmds[6]
		cli.Send(from, to, amount, miner, data)
	case "createWallet":
		fmt.Printf("创建钱包\n")
		cli.CreateWallet()
	case "listAddresses":
		fmt.Printf("打印所有钱包地址\n")
		cli.ListAddresses()
	case "printTransaction":
		fmt.Printf("打印所有交易\n")
		cli.PrintTransaction()
	default:
		fmt.Printf("无用命令！！")
		fmt.Printf(usage)

	}
}

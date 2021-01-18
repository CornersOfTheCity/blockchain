package main

import (
	"blockabout/base58"
	"blockabout/bolt"
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"log"
	"os"
)

type BlockChain struct {
	db   *bolt.DB //句柄
	tail []byte   //最后一个区块的哈希
}

//定义一个UTXOInfo结构，用以找到所有的output和output定位
type UTXOInfo struct {
	TXID   []byte   //交易ID
	Index  int64    //output的索引值
	Output TXOutput //符合要求的output

}

//区块链迭代器
type BlockChainIterator struct {
	db      *bolt.DB
	current []byte
}

const genesisInfo = "这是一个创世块"
const blockChainDB = "blockChain.db"
const blockBucketName = "blockBucket"
const lastHashkey = "lastHashkey"

//创建一个区块链
func CreateBlockChain(miner string) *BlockChain {

	if IsFileExist(blockChainDB) {
		fmt.Printf("区块链已经存在，不需要重复创建\n")
		return nil
	}

	//读写方式打开数据库
	db, err := bolt.Open(blockChainDB, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	//defer db.Close()

	var tail []byte

	//判断是否存在bucket，没有则创建
	db.Update(func(tx *bolt.Tx) error {
		bu, err := tx.CreateBucket([]byte(blockBucketName))
		if err != nil {
			log.Panic(err)
		}
		//开始添加创世块
		//创世块中只有一个挖矿交易
		coinbase := NewCoinBaseTx(miner, genesisInfo)
		genesisBlock := NewBlock([]*Transaction{coinbase}, []byte{})
		bu.Put(genesisBlock.Hash, genesisBlock.Serialize())
		bu.Put([]byte(lastHashkey), genesisBlock.Hash)

		tail = genesisBlock.Hash
		return nil
	})
	return &BlockChain{db, tail}
}

//返回区块链实例
func NewBlockChain() *BlockChain {
	//genesisBlock := NewBlock("genesisInfo", []byte{0x00000000000000})
	//bc := BlockChain{Blocks: []*Block{genesisBlock}}
	//return &bc

	if !IsFileExist(blockChainDB) {
		fmt.Printf("区块链不存在，请先创建\n")
		return nil
	}

	//读写方式打开数据库
	db, err := bolt.Open(blockChainDB, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	//defer db.Close()

	var tail []byte

	//判断是否存在bucket，没有则创建
	db.View(func(tx *bolt.Tx) error {
		bu := tx.Bucket([]byte(blockBucketName))
		if bu == nil {
			fmt.Printf("区块链bucket不存在，请检查\n")
			os.Exit(0)

		} else {
			tail = bu.Get([]byte(lastHashkey))
		}
		return nil
	})
	return &BlockChain{db, tail}
}

func (bc *BlockChain) AddBlock(txs []*Transaction) {
	//矿工得到交易时，第一时间对交易进行验证
	validTXs := []*Transaction{}
	for _, tx := range txs {
		if bc.VerifyTransaction(tx) {
			fmt.Printf("有效交易：%x\n", tx.TXId)
			validTXs = append(validTXs, tx)
		} else {
			fmt.Printf("发现无效交易：%x\n", tx.TXId)
		}
	}
	bc.db.Update(func(tx *bolt.Tx) error {
		bu := tx.Bucket([]byte(blockBucketName))
		if bu == nil {
			fmt.Printf("bucket不存在，请检查！\n")
			os.Exit(1)
		}
		block := NewBlock(txs, bc.tail)
		bu.Put(block.Hash, block.Serialize())
		bu.Put([]byte(lastHashkey), block.Hash)
		bc.tail = block.Hash
		return nil
	})
}

//创建迭代器并初始化
func (bc *BlockChain) NewIterator() *BlockChainIterator {
	return &BlockChainIterator{bc.db, bc.tail}
}

func (it *BlockChainIterator) Next() *Block {
	var block Block
	it.db.View(func(tx *bolt.Tx) error {
		bu := tx.Bucket([]byte(blockBucketName))
		if bu == nil {
			fmt.Printf("bucket不存在，请检查\n")
			os.Exit(2)
		}

		blockInfo := bu.Get(it.current)
		block = *Deserialize(blockInfo)

		it.current = block.PrevBlockHash
		return nil
	})
	return &block
}

//找到所有的UTXO
func (bc *BlockChain) FindMyUtxos(pubKeyHash []byte) []UTXOInfo {

	var UTXOInfos []UTXOInfo

	//标识已经消耗过的UTXO结构，key为交易ID，value是这个ID里面的output索引的数组
	spentUTXOs := make(map[string][]int64)

	//使用迭代器遍历账本
	it := bc.NewIterator()
	for {
		block := it.Next()
		//遍历交易
		for _, tx := range block.Transactions {

			if tx.IsCoinbase() == false {
				//遍历input
				for _, input := range tx.TXInputs {

					//找到属于自己所有的input（自己转给别人的当作一个input，address表示自己可以解锁）
					//判断当前被使用的input是否为目标地址所有
					if bytes.Equal(hashPubKey(input.PubKey), pubKeyHash) {
						fmt.Printf("找到消耗过的的output！index:%d\n", input.index)
						key := string(input.TXID)
						spentUTXOs[key] = append(spentUTXOs[key], input.index)
					}

				}
			}
			//这笔交易中消耗过的output
			key := string(tx.TXId)
			indexes := spentUTXOs[key]
		OUTPUT:
			//遍历output
			for i, output := range tx.TXOutputs {
				if len(indexes) != 0 {
					fmt.Printf("这笔交易中有被消耗的output!\n")
					for _, j := range indexes {
						if int64(i) == j {
							fmt.Printf("i==j,当前output已经被消耗过，跳过不统计\n")
							continue OUTPUT //跳过此次遍历
						}
					}
				}

				//找到属于自己所有的output（别人转给自己的当作一个output，address依旧暂时表示自己可以解锁）
				if bytes.Equal(pubKeyHash, output.PubKeyHash) {
					//fmt.Printf("找到了属于%s的output，i：%d\n", address, i)
					utxoinfo := UTXOInfo{tx.TXId, int64(i), output}
					UTXOInfos = append(UTXOInfos, utxoinfo)
				}

			}
		}

		if len(block.PrevBlockHash) == 0 {
			fmt.Printf("遍历结束\n")
			break
		}
	}
	//return []UTXOInfo{}
	return UTXOInfos
}

func (bc *BlockChain) GetBalance(address string) {

	//这个过程不要打开钱包，因为可能查看余额的人不是地址本人
	decodeInfo, err := base58.Decode(address)
	if err != nil {
		fmt.Printf("解码出错！\n")
		log.Panic(err)
	}

	//从25个字节中截取其中的20个得到公钥哈希
	pubKeyHash := decodeInfo[1 : len(decodeInfo)-4]
	utxoinfos := bc.FindMyUtxos(pubKeyHash)
	var total = 0.0
	for _, utxoinfo := range utxoinfos {
		total += utxoinfo.Output.Value
	}
	fmt.Printf("%s的余额为%f\n", address, total)

}

//遍历账本，找到属于付款人的合适金额，然后把这个outputs找到
func (bc *BlockChain) FindNeedUtxos(pubKeyHash []byte, amount float64) (map[string][]int64, float64) {

	needutxos := make(map[string][]int64)

	var resValue float64 //统计的金额

	//复用findmuutxo函数
	utxoinfos := bc.FindMyUtxos(pubKeyHash)
	for _, utxoinfo := range utxoinfos {
		key := string(utxoinfo.TXID)
		needutxos[key] = append(needutxos[key], int64(utxoinfo.Index))
		resValue += utxoinfo.Output.Value

		if resValue >= amount {
			break
		}
	}
	return needutxos, resValue
}

//签名交易
func (bc *BlockChain) SignTransaction(tx *Transaction, privateKey *ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	//遍历tx的inputs，通过ID去查找所引用的交易
	for _, input := range tx.TXInputs {
		prevTx := bc.FindTransaction(input.TXID)
		if prevTx == nil {
			fmt.Printf("没有找到交易:%x\n", input.TXID)
		} else {
			//把找到的引用交易保存起来
			prevTXs[string(input.TXID)] = *prevTx
		}
	}
	tx.Sign(privateKey, prevTXs)
}

//矿工校验流程
//1。找到交易input所引用的所有交易prevTXs
//2。对交易进行验证
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {

	//挖矿交易直接返回true
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)
	//遍历tx的inputs，通过ID去查找所引用的交易
	for _, input := range tx.TXInputs {
		prevTx := bc.FindTransaction(input.TXID)
		if prevTx == nil {
			fmt.Printf("没有找到交易:%x\n", input.TXID)
		} else {
			//把找到的引用交易保存起来
			prevTXs[string(input.TXID)] = *prevTx
		}
	}
	return tx.Verify(prevTXs)
}

func (bc *BlockChain) FindTransaction(txid []byte) *Transaction {
	//遍历区块链的交易
	//通过对比id来识别
	it := bc.NewIterator()
	for {
		block := it.Next()
		//如果找到相同ID的交易则直接返回交易
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.TXId, txid) {
				fmt.Printf("找到了所引用的交易：%x\n", tx.TXId)
				return tx
			}
			if len(block.PrevBlockHash) == 0 {
				break
			}
		}
	}
	return nil
}

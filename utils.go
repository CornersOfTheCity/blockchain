//工具函数文件
package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

//用来将uint转化为byte
func uintToByte(num uint64) []byte {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buffer.Bytes()
}

//判断文件是否存在
func IsFileExist(fileName string) bool {
	//使用os.stat来判断
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

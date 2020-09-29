package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"test/MetaData"
	"test/config"
	mq "test/msgQue"
	"test/mysql"
	"test/oss"
	"test/transfor"
)

//作用是启动消息队列生产者和消费者
func ProcessTransfer(msg []byte) bool {
	//1.解析msg
	//2.根据临时存储文件路径，创建文件句柄
	//3.通过文件句柄将文件内容读出并且上传
	//4.更新文件存储路径表

	log.Println("收到消息")
	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, &pubData)

	if err != nil {
		log.Println(err)
		return false
	}
	filed, err := os.Open(pubData.CurLocation)
	if err != nil {
		log.Println(err)
		return false
	}

	err = oss.Bucket().PutObject(
		pubData.DestLocation,
		bufio.NewReader(filed))
	if err != nil {
		log.Println(err)
		return false
	}

	db := mysql.DBConn()
	//修改数据库
	transfor.InsertSingleOssRecord(db, MetaData.FileMeta{
		FileSha1: pubData.FileHash,
		Location: pubData.DestLocation,
	})
	return true
}

func main() {
	log.Println("开始监听")
	suc := mq.StartConsume(
		config.TransOSSErrQueueName,
		"transfer_oss",
		ProcessTransfer)
	fmt.Println(suc)
}

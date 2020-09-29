package msgQue

import (
	"fmt"
	"log"
	"test/config"

	"github.com/streadway/amqp"
)

//没有私有云，所以不要common
type TransferData struct {
	FileHash     string
	CurLocation  string
	DestLocation string
}

var conn *amqp.Connection
var channel *amqp.Channel

func initChannel() bool {
	//1.判断channel是否创建过
	if channel != nil {
		return true
	}
	//2.获取链接
	conn, err := amqp.Dial(config.RabbitmqURL)
	if err != nil {
		fmt.Println(err)
		return false
	}
	//3.打开ch用于消息的发布与接收
	channel, err = conn.Channel()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func Publish(exchange, routingKey string, msg []byte) bool {
	//1.检查channel
	if !initChannel() {
		return false
	}
	//2.调用ch的publish方法
	err := channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func StartConsume(qName, cName string, callback func([]byte) bool) bool {
	if initChannel() == false {
		log.Println("Channel can't up")
		return false
	}
	//1.获取消息信道
	msgs, err := channel.Consume(
		qName,
		cName,
		true,
		false, //竞争机制
		false,
		false,
		nil,
	)

	if err != nil {
		fmt.Println(err)
		return false
	}

	done := make(chan bool)

	//2.循环获取队列消息
	go func() {
		for msg := range msgs {
			procssSuc := callback(msg.Body)
			//3.调用callback函数进行处理
			if !procssSuc {
				//to do
			}
		}
	}()

	<-done
	return true
}

package database

import (
	"encoding/json"
	"fmt"
	"log"

	"MyGoProject/internal/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

var MQChannel *amqp.Channel

func StartConsume() {
	msgs, err := MQChannel.Consume(
		"resume_queue",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("获取消费通道失败:%v", err)
	}
	go func() {
		for d := range msgs {
			fmt.Printf("收到新任务:%s\n", d.Body)
			success := processTask(d.Body)
			if success {
				d.Ack(false)
			} else {
				d.Nack(false, true)
			}
		}
	}()
	fmt.Println("消费者已启动,等待中")
}

func processTask(body []byte) bool {
	var msg model.ResumeMessage
	err := json.Unmarshal(body, &msg)
	if err != nil {
		fmt.Println("解析JSON错误:", err)
		return true
	}
	newResume := model.Resume{
		UserID:  msg.UserID,
		JobID:   msg.JobID,
		Content: msg.Content,
		Status:  1,
	}
	if err := DB.Create(&newResume).Error; err != nil {
		log.Printf("存入数据库失败:%v", err)
		return false
	}
	log.Printf("简历 [用户:%d -> 岗位:%d] 异步写入成功!", msg.UserID, msg.JobID)
	return true
}

func InitMQ() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("连接RabbitMQ失败:%v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("打开MQ通道失败:%v", err)
	}
	_, err = ch.QueueDeclare(
		"resume_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("声明队列失败:%v", err)
	}
	MQChannel = ch
	fmt.Println("RabbitMQ连接并初始化成功")
}

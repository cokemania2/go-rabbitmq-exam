package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

var amqpConn *amqp.Connection

type publishData struct {
	Command string `json:"command"`
	Group   string `json:"group"`
	Topic   string `json:"topic"`
}

func main() {
	//connect rabbitmq connection
	AmqpConnectionSet("127.0.0.1")
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		PublishTo("testName", "", []byte(`{"command":"pong"}`))
	})
	r.GET("/publish/:group", func(c *gin.Context) {
		data, err := json.Marshal(publishData{Command: "publishGroup", Group: c.Param("group")})
		checkErr(err)
		PublishTo("testName", "", data)
	})
	r.GET("/publish/:group/*topic", func(c *gin.Context) {
		data, err := json.Marshal(publishData{Command: "publishTopic", Group: c.Param("group"), Topic: strings.Split(c.Param("topic"), "/")[1]})
		checkErr(err)
		PublishTo("testName", "", data)
	})
	r.Run()
}

func AmqpConnectionSet(MQServerIP string) {
	var err error
	fmt.Println("amqp://test:test@" + MQServerIP + ":5672/")
	amqpConn, err = amqp.Dial("amqp://test:test@" + MQServerIP + ":5672/")
	checkErr(err)
	//connect rabbitmq channel

}

func PublishTo(exchangeName string, routingKey string, marshaledData []byte) {
	amqpChannel, err := amqpConn.Channel()
	checkErr(err)
	err = amqpChannel.Publish(
		exchangeName,
		routingKey, // channelID
		false,      //mandatory
		false,      //immediate
		amqp.Publishing{
			Headers:      amqp.Table{},
			ContentType:  "application/json",
			Body:         marshaledData,
			DeliveryMode: amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:     0,              // 0-9
		},
	)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

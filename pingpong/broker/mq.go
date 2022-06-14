package broker

import (
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

var (
	AmqpConn    *amqp.Connection
	AmqpChannel *amqp.Channel
	Deliveries  <-chan amqp.Delivery
	SocketHub   *Hub
	AmqpQueue   *amqp.Queue
)

func queueDeclare(channel *amqp.Channel, queueName string) (*amqp.Queue, error) {
	if channel != nil {
		queue, err := channel.QueueDeclare(
			queueName, // name of the queue
			true,      // durable
			true,      // delete when unused
			false,     // exclusive
			false,     // noWait
			nil,       // arguments
		)
		if err == nil {
			return &queue, err
		}
	}
	return &amqp.Queue{}, fmt.Errorf("err")
}
func exchangeDeclare(channel *amqp.Channel, exchangeName string, exchangeType string) {
	if channel != nil {
		err := channel.ExchangeDeclare(
			exchangeName, // name of the exchange
			exchangeType, // type
			false,        // durable
			false,        // delete when complete
			false,        // internal
			false,        // noWait
			nil,          // arguments
		)
		fmt.Println(err)
	}
}
func bindQueue(channel *amqp.Channel, queueName string, exchangeName string, key string) {
	err := channel.QueueBind(
		queueName,    // name of the queue
		key,          // bindingKey
		exchangeName, // sourceExchange
		false,        // noWait
		nil,          // arguments
	)
	checkErr(err)
}
func PublishTo(channel *amqp.Channel, exchangeName string, routingKey string, marshaledData []byte) {
	err := channel.Publish(
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
func StartConsume() {
	hostName, err := os.Hostname()
	checkErr(err)
	AmqpQueue, err = queueDeclare(AmqpChannel, hostName+" test ")
	checkErr(err)
	exchangeDeclare(AmqpChannel, "mailplug", "fanout")
	bindQueue(AmqpChannel, AmqpQueue.Name, "mailplug", "")
	err = AmqpChannel.Qos(0, 0, false)
	checkErr(err)
	Deliveries, err = AmqpChannel.Consume(
		AmqpQueue.Name, // name
		hostName,       // consumerTag,
		false,          // autoAck
		false,          // exclusive
		false,          // noLocal
		false,          // noWait
		nil,            // arguments
	)
	checkErr(err)

}
func ConnectAMQP(serverIP string) {
	var err error
	AmqpConn, err = amqp.Dial("amqp://test:test@" + serverIP + "/")
	checkErr(err)
	AmqpChannel, err = AmqpConn.Channel()
	checkErr(err)
}
func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

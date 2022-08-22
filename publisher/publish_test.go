package main

import (
	"testing"
)

var (
	Environment = "development"
)

func TestConnection(t *testing.T) {
	// err := godotenv.Load("/home/ubuntu/puddlr/api3/.env." + Environment)
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }
	AmqpConnectionSet("127.0.0.1")
	PublishTo("testExchange", "testRoutingKey", []byte(`{"command":"test"}`))
}

// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Multicast struct {
	group string
	topic string
	data  []byte
}

type publishData struct {
	Publish string `json:"publish"`
	Group   string `json:"group"`
	Topic   string `json:"topic"`
}

type Hub struct {
	// Registered clients.
	Groups

	// Inbound messages from the clients.
	broadcast chan []byte

	multicastGroup chan Multicast

	multicastTopic chan Multicast

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

var funcMap = map[string]interface{}{
	"publishGroup": publishGroup,
	"publishTopic": publishTopic,
	"pong":         pong,
}

func NewHub() *Hub {
	return &Hub{
		multicastTopic: make(chan Multicast),
		multicastGroup: make(chan Multicast),
		broadcast:      make(chan []byte),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		Groups:         make(Groups),
	}
}

func (h *Hub) Run() {
	StartConsume()
	for {
		select {
		// on_open
		case client := <-h.register:
			if _, ok := h.Groups[client.group]; !ok {
				h.Groups[client.group] = Topic{client.topic: make(Clients)}
			} else if _, ok := h.Groups[client.group][client.topic]; !ok {
				h.Groups[client.group][client.topic] = make(Clients)
			}
			h.Groups[client.group][client.topic][client] = true
		// on_close
		case client := <-h.unregister:
			if _, ok := h.Groups[client.group][client.topic][client]; ok {
				delete(h.Groups[client.group][client.topic], client)
				close(client.send)
			}
		case mqMessage := <-Deliveries:
			message := mqMessage.Body
			var data map[string]string
			json.Unmarshal(message, &data) // message to JSON TYPE
			message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
			if command, ok := funcMap[data["command"]]; ok {
				command.(func([]uint8))(message)
			} else {
				fmt.Println(data["command"] + " is not a func.")
			}
			mqMessage.Ack(false)

		case message := <-h.broadcast:
			for _, groups := range h.Groups {
				for _, topics := range groups {
					for client := range topics {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(topics, client)
						}
					}
				}
			}
		case message := <-h.multicastGroup:
			for _, topics := range h.Groups[message.group] {
				for client := range topics {
					select {
					case client.send <- message.data:
					default:
						close(client.send)
						delete(topics, client)
					}
				}
			}
		case message := <-h.multicastTopic:
			for client := range h.Groups[message.group][message.topic] {
				select {
				case client.send <- message.data:
				default:
					close(client.send)
					delete(h.Groups[message.group][message.topic], client)
				}
			}
		}
	}
}

func pong(data []uint8) {
	go func() {
		SocketHub.broadcast <- data
	}()
}

func publishGroup(data []uint8) {
	var d publishData
	json.Unmarshal(data, &d)
	go func() {
		SocketHub.multicastGroup <- Multicast{d.Group, "", data}
	}()
}

func publishTopic(data []uint8) {
	var d publishData
	json.Unmarshal(data, &d)
	go func() {
		SocketHub.multicastTopic <- Multicast{d.Group, d.Topic, data}
	}()
}

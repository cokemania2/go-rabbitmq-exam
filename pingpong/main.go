// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"net/http"

	"example.com/broker"
	"github.com/gin-gonic/gin"
)

// var addr2 = flag.String("addr", ":8082", "http service address")

func serveHome(c *gin.Context) {
	parm := c.Param("group")
	c.HTML(http.StatusOK, "home.html", gin.H{
		"group": parm,
	})
}

func main() {

	flag.Parse()
	broker.ConnectAMQP("127.0.0.1")
	broker.SocketHub = broker.NewHub()
	go broker.SocketHub.Run()
	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	router.LoadHTMLGlob("home.html")
	router.GET("/consume/:group", serveHome)
	router.GET("/consume/:group/*topic", serveHome)
	router.GET("/ws/:group", broker.ServeWs)
	router.GET("/ws/:group/*topic", broker.ServeWs)
	router.Run(":8082")
}

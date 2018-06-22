// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"log"
	"net/url"
	"github.com/gorilla/websocket"
	"time"
)

var addr = flag.String("addr", "localhost:7777", "http service address")

func loop() {
	for {
		u := url.URL{Scheme: "ws", Host: *addr, Path: "/connect"}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			continue
		}
		// 循环读消息
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				// log.Println("read:", err)
				break
			}
			// log.Printf("recv: %s", message)
		}
		c.Close()
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	for i := 0; i < 10000; i++ {
		go loop()
	}

	for {
		time.Sleep(1 * time.Second)
	}
}
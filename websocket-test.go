package main

import (
	"fmt"
	"flag"
	"gographer/gowebsocket"
	"log"
	"strconv"
	"time"
)

// Usage example
func main() {
	// cmd line flags
	fmt.Println(">>>>> before <<<<")
	var ip = flag.String("ip", "127.0.0.1", "ip to run on")
	var port = flag.String("port", ":3999", "port to run on")
	flag.Parse()
	fmt.Println(">>>>> after <<<<")


	// Start Server
	server := gowebsocket.New(*ip, *port)
	server.Start()

	// Start a single client which sends and receives messages
	c, err := gowebsocket.NewClient(*ip, *port)
	if err != nil {
		log.Panic("Unable to start WS Client with err: ", err)
	}
	i := 0
	for {
		time.Sleep(1 * time.Second)

		msg := "Hello, Websocket. This is message " + strconv.Itoa(i)

		log.Printf("Sending: %s \n", msg)

		c.Send(msg)
		recv := c.Receive()

		log.Printf("Received: %s\n", recv)
		i++
	}

}

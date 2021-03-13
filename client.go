package gowebsocket

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net"
	"time"
)

type Client struct {
	Conn     *websocket.Conn
	Listener net.Listener
}

func NewClient(ip, port string) (c *Client, err error) {
	c = new(Client)
	address := "ws://" + ip + port
	origin := "http://" + ip

	var retryAttempts int = 0
	for {
		c.Conn, err = websocket.Dial(address, "ws", origin)
		if err != nil {
			if retryAttempts == 10 {
				return nil, err
			}
			retryAttempts++
			time.Sleep(125 * time.Millisecond)
			continue
		}
		break
	}
	return c, nil
}

func (c *Client) Send(message string) {
	msg := []byte(message)
	c.SendBytes(msg)
}

func (c *Client) SendBytes(message []byte) {
	if _, err := c.Conn.Write(message); err != nil {
		log.Panic("Message", message, " could not be sent", err)
	}

}
func (c *Client) Receive() string {

	var incoming = make([]byte, 512)
	var n int
	var err error

	if n, err = c.Conn.Read(incoming); err != nil {
		log.Fatal(err)
	}
	return string(incoming[:n])

}

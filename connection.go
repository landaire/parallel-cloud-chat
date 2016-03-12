package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	messageSize = 1024

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var pool = connectionPool{
	broadcast:   make(chan []byte, 256),
	connections: make(map[connection]bool),
	register:    make(chan *connection),
	deregister:  make(chan *connection),
}

type connection struct {
	username string
	ws       *websocket.Conn
	messages chan []byte
}

type connectionPool struct {
	broadcast   chan []byte
	connections map[connection]bool
	register    chan *connection
	deregister  chan *connection
}

func (c *connection) readMessages() {
	defer func() {
		pool.deregister <- c
		c.ws.Close()
	}()

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
			fmt.Fprintln(os.Stderr, "error:", err)
			break
		}

		pool.broadcast <- message
		go insertMessage(c.username, string(message))
	}
}

func (c *connection) sendMessages() {
	pool.register <- c

	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.messages:
			// If the channel is closed then close the connection with the client
			// in this case, the server might be going down or something
			if !ok {
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Try sending the message. If this fails then the client might have
			// already closed the connection
			if err := c.ws.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			// Each tick we send a ping to the client to make sure the connection
			// is still alive
			if err := c.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

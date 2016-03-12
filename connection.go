package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	maxMessageSize = 1024

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var pool = connectionPool{
	broadcast:   make(chan message, 256),
	connections: make(map[*connection]bool),
	register:    make(chan *connection),
	deregister:  make(chan *connection),
}

type connection struct {
	username string
	ws       *websocket.Conn
	messages chan []byte
}

type connectionPool struct {
	broadcast   chan message
	connections map[*connection]bool
	register    chan *connection
	deregister  chan *connection
}

type message struct {
	username string
	message  []byte
}

func (p *connectionPool) run() {
	for {
		select {
		case c := <-p.register:
			p.connections[c] = true
		case c := <-p.deregister:
			delete(p.connections, c)
		case m := <-p.broadcast:
			for c := range p.connections {
				select {
				case c.messages <- []byte(fmt.Sprintf("%s:\n%s", m.username, string(m.message))):
				default:
					close(c.messages)
					delete(p.connections, c)
				}
			}
		}
	}
}

func (c *connection) readMessages() {
	defer func() {
		pool.deregister <- c
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		t, m, err := c.ws.ReadMessage()
		if err != nil && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
			fmt.Fprintln(os.Stderr, "error:", err)
			return
		}

		if t == websocket.TextMessage {
			fmt.Println("Got a message from ", c.username)
			pool.broadcast <- message{username: c.username, message: m}
			insertMessage(c.username, string(m))
		}
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
			fmt.Println("Sending message to user", c.username)

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

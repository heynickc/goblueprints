package main

import (
	"github.com/gorilla/websocket"
)

// client represents a single chatting user
type client struct {
	// socket is a web socket for this client
	socket *websocket.Conn
	// send is a channel on which messages are sent
	// it is a buffered channel through which received messages
	// are queued ready to be forwarded to the user's browser
	// via a websocket
	send chan []byte
	// room is the room this client is chatting in
	room *room
}

// allows the client to read from the socket via ReadMessage
// it sends recieved messages from the browser to the
// forward channel on the room type to show up in the room
// that the client has open in their browser
func (c *client) read() {
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

// accepts messages from the send channel, the mesages that the user enters
// are written to the socket to be displayed on everyone's browser session
func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close()
}

// at any time, the socket fails, the loop is broken and the socket is closed

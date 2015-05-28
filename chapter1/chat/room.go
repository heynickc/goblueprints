package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/heynickc/goblueprints/chapter1/trace"
)

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients
	forward chan []byte
	// join is a channel for clients wishing to join the room
	join chan *client
	// leave is a channel for clients wishing to leave the room
	leave chan *client
	// clients holds all the current clients in this room
	clients map[*client]bool
	// tracer will reveive trace information of activity
	// in teh room
	tracer trace.Tracer
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	// for loop indicates that this method will run forever
	// unless the program is terminated
	// it will run as a goroutine - in the background and won't block
	// the rest of the application
	for {
		// if the message is recieved on any of these channels
		// the select statement will run the code for that case
		// this is how we are able to synchronize to ensure that our
		// r.clients map is only ever modified by one thing at a time
		select {
		case client := <-r.join:
			// joining
			// we update the room's clients map
			// adding a reference to the client as the key
			// and true - pretty much a slice
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			// leaving
			// do the opposite of above, delete the client
			// from the map
			delete(r.clients, client)
			// closing the
			close(client.send)
			r.tracer.Trace("Client left")
		case msg := <-r.forward:
			r.tracer.Trace("Message received: ", string(msg))
			// forward the message to all clients
			// iterate over all the clients
			// send a message down each of their sockets
			for client := range r.clients {
				select {
				case client.send <- msg:
					// send the message
					// this is picked up by the client.write() method
					// which is displayed on the client's browser
					r.tracer.Trace(" -- sent to the client")
				default:
					// failed to send this means the client needs
					// to be taken out of the room just to clean things up
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace(" -- failed to send, cleaned up the client")
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

// an upgrader takes a normal http connection
// and upgrades it to make it able to use web sockets
// we only need to create one and point to it because it's reusable
var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

// we have satisfied the interface to be an http handler
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// when a request enters
	// we upgrade it to get the socket
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		// I forgot to add this and it still worked,
		// need to look into why
		return
	}
	// create a client and pass it
	// to the join channel for the room
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	// send the client into the room channel
	r.join <- client
	// we always want to leave the room after everything?
	defer func() { r.leave <- client }()
	// runs asynchronously, writing to the socket
	go client.write()
	// not sure what's going on here...
	// it's used to block the main thread until
	// it's time to close it
	client.read()
}

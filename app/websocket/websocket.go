package websocket

import (
	"fmt"
	"sync"
)

// hub maintains the set of active clients and broadcasts messages to the
// clients.
type hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
	once       *sync.Once
	canDelete  chan<- string
	key        string
}

func newhub(canDelete chan<- string) *hub {
	return &hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		canDelete:  canDelete,
		once:       &sync.Once{},
	}
}

func (h *hub) run() {
	h.once.Do(func() {
		for {
			select {
			case client := <-h.register:
				h.clients[client] = true
			case client := <-h.unregister:
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
					if l := len(h.clients); l == 0 {
						fmt.Println(l, " elements notify")
						h.canDelete <- h.key
					} else {
						fmt.Println(l, " elements")
					}
				}
			case message := <-h.broadcast:
				for client := range h.clients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	})
}

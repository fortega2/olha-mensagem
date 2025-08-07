package websocket

type Hub struct {
	clients    map[*Client]struct{}
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	shutdown   chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		shutdown:   make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
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
		case <-h.shutdown:
			return
		}
	}
}

func (h *Hub) Shutdown() {
	close(h.shutdown)

	for client := range h.clients {
		close(client.send)
		client.conn.Close()
		delete(h.clients, client)
	}

	close(h.broadcast)
	close(h.register)
	close(h.unregister)
}

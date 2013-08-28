package engine

type Hub struct {
  // Registered connections.
  connections map[*Connection]bool

  // Inbound messages from the connections.
  broadcast chan GameObject

  // Register requests from the connections.
  register chan *Connection

  // Unregister requests from connections.
  unregister chan *Connection
}

func NewHub() (*Hub) {
  return &Hub{
    broadcast:   make(chan GameObject),
    register:    make(chan *Connection),
    unregister:  make(chan *Connection),
    connections: make(map[*Connection]bool),
  }
}

func (h *Hub) Run() {
  for {
    select {
    case c := <-h.register:
      h.connections[c] = true
    case c := <-h.unregister:
      delete(h.connections, c)
      close(c.send)
    case m := <-h.broadcast:
      for c := range h.connections {
        select {
        case c.send <- m:
        default:
          delete(h.connections, c)
          close(c.send)
          go c.ws.Close()
        }
      }
    }
  }
}


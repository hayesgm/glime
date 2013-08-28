package engine

import (
  "log"
  "code.google.com/p/go.net/websocket"
  "encoding/json"
)

type Connection struct {
  // The websocket connection.
  ws *websocket.Conn

  // Id of player object
  id int

  // Buffered channel of outbound messages.
  send chan GameObject
}

func (c *Connection) Reader(ch chan GameMessage) {
  for {
    var message GameMessage
    err := websocket.JSON.Receive(c.ws, &message)
    if err != nil {
      log.Printf("Failed to read message, %s\n", err)
       break
    }
    message.id = c.id
    log.Printf("Got message: %#v\n", message)
    ch <- message
  }
  c.ws.Close()
}

func (c *Connection) Writer() {
  for message := range c.send {
    m, _ := json.Marshal(&message)
    log.Printf("Sending message: %#v, %s", message, m)
    err := websocket.JSON.Send(c.ws, message)
    if err != nil {
      log.Printf("Failed to send message, %s\n", err)
    }
    log.Printf("Sent message: %#v\n", message)
  }
  c.ws.Close()
}

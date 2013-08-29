package engine

import (
  "net/http"
  "code.google.com/p/go.net/websocket"
  "log"
  "encoding/json"
  "math"
  "time"
)

var chGame = make(chan GameMessage)

type GameMessage struct {
  Type string
  Payload string
  id int // Which player this was received from
}

type MoveMessage struct {
  Direction string
  Bearing float64
}

type FireMessage struct {
  Bearing float64
}

func processGameMessage(message GameMessage) {
  // This function is responsible for updating the game object
  // GameState is going to need to track deltas and send delta'd objects back
  // to all receivers
  log.Printf("Processing game message: %#v", message)
  obj := game.objects[message.id]
  if obj == nil {
    log.Println("Failed to find Player: ", message.id)
    return
  }

  switch message.Type {
  case "move":
    var move MoveMessage
    var speed = (*obj.Characteristics)["speed"]

    err := json.Unmarshal([]byte(message.Payload), &move)
    if err != nil {
      log.Println("Failed to parse move message:", err)
      return
    }

    go func() {
      var accel = [6]float64{0.2,0.3,0.5,0.8,0.7,0.2}

      for i := 0; i < len(accel); i++ { // Add multiple movements here
        <- obj.lock
        
        // log.Printf("Updating obj, speed: %v, bearing: %v\n", speed, move.Bearing)
        oX, oY := obj.X, obj.Y // Store originals

        switch (move.Direction) {
        case "toward":
          obj.X += speed * math.Cos(move.Bearing) * accel[i]
          obj.Y += speed * math.Sin(move.Bearing) * accel[i]
        case "away":
          obj.X -= speed * math.Cos(move.Bearing) * accel[i]
          obj.Y -= speed * math.Sin(move.Bearing) * accel[i]
        case "strafe-right":
          obj.X += speed * math.Sin(move.Bearing) * accel[i]
          obj.Y += speed * math.Cos(move.Bearing) * accel[i]
        case "strafe-left":
          obj.X -= speed * math.Sin(move.Bearing) * accel[i]
          obj.Y -= speed * math.Cos(move.Bearing) * accel[i]
        }
        
        obj.Bearing = move.Bearing
        obj.lock <- 1 // Since collision detection could be slow, we'll release the lock as that runs

        collisions := obj.collisions(game, obj)

        for _, v := range collisions {
          switch (v.ObjType) {
          case "Player":
            <- obj.lock
            obj.X, obj.Y = oX, oY // Restore originals
            obj.lock <- 1
          case "Projectile":
            // He's dead, jim.
            obj.Respawn()
          }
        }
        
        game.update <- obj

        time.Sleep(30*time.Millisecond)
      }
    }()
  case "fire":
    var fire FireMessage
    
    err := json.Unmarshal([]byte(message.Payload), &fire)
    if err != nil {
      log.Println("Failed to parse fire message:", err)
      return
    }

    player := game.objects[message.id]

    log.Printf("Creating new projectile from (%v,%v)", player.X, player.Y)

    qualities := make(map[string]string)
    qualities["color"] = (*player.Qualities)["color"]

    characteristics := make(map[string]float64)
    characteristics["speed"] = 10.0
    characteristics["width"] = 1.0

    obj := game.NewObject(player.Id, player.X, player.Y, fire.Bearing, "Projectile", &qualities, &characteristics)
    game.add <- obj

    // This will actually fire the Projectile
    go func() {
      for i := 0; i < 100; i++ {
        <- obj.lock
        obj.X += (*obj.Characteristics)["speed"] * math.Cos(fire.Bearing)
        obj.Y += (*obj.Characteristics)["speed"] * math.Sin(fire.Bearing)
        obj.Bearing = fire.Bearing
        obj.lock <- 1
        game.update <- obj

        // check for a collision with a player
        collisions := obj.collisions(game, player)

        for _, v := range collisions {
          switch (v.ObjType) {
          case "Player":
            // He's dead, jim.
            v.Respawn()
            game.update <- v
          }
        }

        time.Sleep(50*time.Millisecond)
      }
      game.remove <- obj
    }()
  default:
    log.Printf("Unknown game message: %#v\n", message)
    return
  }
}

// This is going to host the core engine code and the websocket
func gameServer(ws *websocket.Conn) {
  log.Printf("Booting game player...%v\n", game)

  obj := CreatePlayer()
  game.add <- obj

  c := &Connection{send: make(chan GameObject, 256), id: obj.Id, ws: ws}
  h.register <- c
  defer func() { h.unregister <- c }()
  go c.Writer()

  c.send <- *obj // First object is player, itself

  for _, obj := range game.objects {
    c.send <- *obj
  }

  c.Reader(chGame)
  // TODO: These reader/writers need to run all reads through processGameMessage
  // Writing will occur when the GameState is updated (possibly via a Tick)
}

func RegisterGameSocket() {
  go func() { // This will process any game messages, as necessary
    for {
      message := <- chGame

      processGameMessage(message)
    }
  }()

  http.Handle("/game", websocket.Handler(gameServer))
  go h.Run()
}

package engine

import (
  "log"
  "math/rand"
  "fmt"
  "math"
)

// The next thing we need to do here is capture the game state, a collection of GameObjects
type GameObject struct {
  Id int
  OwnerId int
  X float64
  Y float64
  Bearing float64
  ObjType string
  Qualities *map[string]string
  Characteristics *map[string]float64
  Remove bool
  lock chan int
  kill chan int
}

// This is the actual game state
type GameState struct {
  objects map[int]*GameObject
  add chan *GameObject
  remove chan *GameObject
  update chan *GameObject
  nextId chan int
}

func (game *GameState) manageGameState() {
  for {
    select {
      case obj := <- game.add:
        log.Printf("Adding object: %#v", obj)
        game.objects[obj.Id] = obj // Add object
        h.broadcast <- *obj
      case obj := <- game.remove:
        log.Printf("Removing object: %#v", obj)
        obj.Remove = true
        h.broadcast <- *obj
        delete(game.objects, obj.Id)
      case obj := <- game.update:
        log.Printf("Updated object: %#v", obj)
        game.objects[obj.Id] = obj
        h.broadcast <- *obj
    }

    log.Printf("Game state: %v", game)
    for k := range game.objects {
      log.Printf("\tGame object: %v->%v", k, game.objects[k])
    }
  }
}

func (game *GameState) processNextId() {
  for i := 0; true; i += 1 {
    game.nextId <- i
  }
}

func NewGame() (game *GameState) {
  game = new(GameState)
  game.objects = make(map[int]*GameObject, 5000) // Max objects!
  game.add = make(chan *GameObject) // Allow buffered options
  game.remove = make(chan *GameObject)
  game.update = make(chan *GameObject)
  game.nextId = make(chan int)
  go game.manageGameState() // This will process insertions / deletions in the game state
  go game.processNextId()

  return
}

func (game *GameState) NewObject(ownerId int, x float64, y float64, bearing float64, objType string, qualities *map[string]string, characteristics *map[string]float64) (obj *GameObject) {
  obj = new(GameObject)
  obj.X = x
  obj.Y = y
  obj.Bearing = bearing
  obj.ObjType = objType
  obj.Qualities = qualities
  obj.Characteristics = characteristics
  obj.Id = <- game.nextId
  obj.OwnerId = ownerId // Can be nil
  obj.lock = make(chan int, 1)
  obj.lock <- 1 // Unlocked
  obj.kill = make(chan int)
  return
}

func randX() float64 {
  return rand.Float64() * 800.0
}

func randY() float64 {
  return rand.Float64() * 600.0
}

func randColor() string {
  return fmt.Sprintf("#%x", int(rand.Float64() * ( 2<<11-1 )))
}

func CreatePlayer() (obj *GameObject) {
  qualities := make(map[string]string)
  qualities["color"] = randColor()

  characteristics := make(map[string]float64)
  characteristics["speed"] = 6.0
  characteristics["width"] = 10.0
  obj = game.NewObject(-1, randX(), randY(), 0.0, "Player", &qualities, &characteristics)

  log.Printf("Created new object: %#v", obj)

  return
}

func (obj *GameObject) distance(other *GameObject) float64 {
  return math.Sqrt( math.Pow( obj.X - other.X, 2) + math.Pow( obj.Y - other.Y, 2 ) )
}

// Checks for collisions with other game objects
// This will ignore any of me or my owned objects
func (obj *GameObject) collisions(game *GameState, self *GameObject) (collisions []*GameObject) {
  collisions = make([]*GameObject, 0)
  
  // log.Printf("Checking collisions for %#v\n", obj)

  for _, v := range game.objects {
    if (v.Id == self.Id || v.OwnerId == self.Id) { // Shouldn't kill ourselves on our own projectiles
      continue
    }
    // log.Printf("\tVerifying (%v v %v) for %#v\n", obj.distance(v), (*obj.Characteristics)["width"] + (*v.Characteristics)["width"], v)
    if obj.distance(v) <= (*obj.Characteristics)["width"] + (*v.Characteristics)["width"] {
      // log.Printf("Collision Detected! %#v <-> %#v\n", obj, v)
      collisions = append(collisions, v)  
    }
  }

  return
}

func (obj *GameObject) Respawn() {
  log.Printf("Killed obj, %#v\n", obj)
  <- obj.lock
  obj.X = randX() // Reboot
  obj.Y = randY()
  (*obj.Qualities)["color"] = randColor()
  (*obj.Characteristics)["speed"] *= 0.8 // Reduce speed on death
  obj.lock <- 1
}

var game = NewGame()
var h = NewHub()

package main

import (
  "net/http"
  "fmt"
  "log"
  "glime/engine"
)

func main() {
  fmt.Println("Running Glime Engine...")
  engine.RegisterStaticAssets("glime.html")
  engine.RegisterGameSocket()
  err := http.ListenAndServe(":1111", nil)
  if err != nil {
    log.Fatal("Failed to launch server", err)
  }
}
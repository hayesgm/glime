package main

import (
  "net/http"
  "fmt"
  "log"
  "github.com/hayesgm/glime/engine"
  "os"
  "path"
)

func main() {
  fmt.Println("Running Glime Engine...")
  wd, err := os.Getwd()
  if (err != nil) {
    log.Fatal("Unable to get cwd", err)
  }
  assetsDir := path.Join(wd, "assets")
  engine.RegisterStaticAssets(assetsDir, "glime.html")
  engine.RegisterGameSocket()
  err = http.ListenAndServe(":1111", nil)
  if err != nil {
    log.Fatal("Failed to launch server", err)
  }
}
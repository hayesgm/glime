package main

import (
  "net/http"
  "fmt"
  "log"
  "github.com/hayesgm/glime/engine"
  "os"
  "path"
  "github.com/coreos/go-etcd/etcd"
  "time"
)

// This is kind of a cheap hack to get our external address
func getAddress() (addr string) {
  resp, err := http.Get("http://instance-data/latest/meta-data/public-ipv4")
  if err != nil {
    return "127.0.0.1"
  }
  buf := make([]byte,256)
  _, err = resp.Body.Read(buf)
  if err != nil {
    log.Fatal("Failed to read local name",err)
  }
  return string(buf)
}

func addToMirrors() {
  // We're going to tell etcd our life
  var c = etcd.NewClient()
  var address = getAddress()

  go func() {
    for {
      log.Println("Adding myself to mirrors",address)
      c.Set(path.Join("mirrors",address),"1",60) // 5 seconds = TTL
      time.Sleep(45*time.Second)
    }
  }()

}
func main() {
  fmt.Println("Running Glime Engine...")
  wd, err := os.Getwd()
  if (err != nil) {
    log.Fatal("Unable to get cwd", err)
  }
  assetsDir := path.Join(wd, "assets")
  engine.RegisterStaticAssets(assetsDir, "glime.html")
  engine.RegisterGameSocket()
  
  addToMirrors()

  err = http.ListenAndServe(":1111", nil)
  if err != nil {
    log.Fatal("Failed to launch server", err)
  }
}
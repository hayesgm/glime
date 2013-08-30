package engine

import (
  "io/ioutil"
  "path"
  "fmt"
  "log"
  "net/http"
  "html/template"
  "github.com/coreos/go-etcd/etcd"
)

var c = etcd.NewClient()

func getHandler(assetsDir, filename string) func(w http.ResponseWriter, req *http.Request) {
  return func(w http.ResponseWriter, req *http.Request) {
    tmpl := template.Must(template.ParseFiles(path.Join(assetsDir,filename)))

    mirrorKeys, err := c.Get("mirrors/") // We're going to grab any server that registers itself as a mirror

    if err != nil {
      log.Println("Failed to reach etcd server:",err)
      tmpl.Execute(w, nil) // Don't require mirrors
    } else {
      mirrors := make([]string,len(mirrorKeys))
      for i, _ := range mirrorKeys {
        mirrors[i] = fmt.Sprintf("http://%s", path.Base(mirrorKeys[i].Key))
      }
      tmpl.Execute(w, mirrors)
    }
  }
}

func RegisterStaticAssets(assetsDir, root string) {
  // We're going to grab each asset from the assets directory and map it by name
  // Plus, we'll map a static route as home
  fmt.Println("Registering Static Assets...")
  files, err := ioutil.ReadDir(assetsDir)
  if err != nil {
    panic(err) // This is serious
  }
  for _, file := range files {
    fmt.Printf("\tRegistering %v\n", file.Name())
    route := fmt.Sprintf("/%s", file.Name())
    
    http.Handle(route, http.HandlerFunc(getHandler(assetsDir, file.Name())))
    if (file.Name() == root) {
      http.Handle("/", http.HandlerFunc(getHandler(assetsDir, file.Name())))
      fmt.Printf("\tRegistered %v to /\n", file.Name())
    }
    fmt.Printf("\tRegistered %v to %v\n", file.Name(), route)
  }
}
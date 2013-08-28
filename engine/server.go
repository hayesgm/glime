package engine

import (
  "io/ioutil"
  "runtime"
  "path"
  "fmt"
  "net/http"
  "html/template"
)

var _, filename, _, _ = runtime.Caller(0)
var assetsDir = path.Join(path.Dir(filename), "../assets")

func getHandler(filename string) func(w http.ResponseWriter, req *http.Request) {
  return func(w http.ResponseWriter, req *http.Request) {
    tmpl := template.Must(template.ParseFiles(path.Join(assetsDir,filename)))
    tmpl.Execute(w, nil)
  }
}

func RegisterStaticAssets(root string) {
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
    
    http.Handle(route, http.HandlerFunc(getHandler(file.Name())))
    if (file.Name() == root) {
      http.Handle("/", http.HandlerFunc(getHandler(file.Name())))
      fmt.Printf("\tRegistered %v to /\n", file.Name())
    }
    fmt.Printf("\tRegistered %v to %v\n", file.Name(), route)
  }
}
package dynstory

import (
    "io/ioutil"
    "net/http"
)

var (
  mainTemplate []byte
)

func init() {
  content, err := ioutil.ReadFile("dynamic-story/mainTemplate.html")
  if err != nil {
    panic(err)
  }
  mainTemplate = content

  http.HandleFunc("/", root)
}

func root(w http.ResponseWriter, r *http.Request) {
  w.Write(mainTemplate)
}
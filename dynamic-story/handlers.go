// Copyright 2016 Philippe Beaudoin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package dynstory

import (
    "appengine"
    "appengine/datastore"
    "appengine/user"
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "html/template"
    "io/ioutil"
    "net/http"
    "strings"
)

type Paragraph struct {
  FullText string
}

type Link struct {
  OutLink string
  FromKey string
  ToKey string
}

var (
  mainTemplate []byte
  editParagraphTemplate = template.Must(template.ParseFiles(
      "dynamic-story/editParagraphTemplate.html"))
)

func init() {
  content, err := ioutil.ReadFile("dynamic-story/mainTemplate.html")
  if err != nil {
    panic(err)
  }
  mainTemplate = content

  http.HandleFunc("/", root)
  http.HandleFunc("/edit-paragraph/", editParagraph)
  http.HandleFunc("/get-paragraphs/", getParagraphs)
  http.HandleFunc("/get-paragraph/", getParagraph)
  http.HandleFunc("/save-paragraph/", saveParagraph)
  http.HandleFunc("/save-link/", saveLink)
}

func root(w http.ResponseWriter, r *http.Request) {
  if !checkUser(w, r) {
    return
  }
  w.Header().Set("Content-type", "text/html; charset=utf-8")
  w.Write(mainTemplate)
}

func editParagraph(w http.ResponseWriter, r *http.Request) {
  if !checkUser(w, r) {
    return
  }
  w.Header().Set("Content-type", "text/html; charset=utf-8")
  err := editParagraphTemplate.Execute(w, r.FormValue("key"))
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func getParagraphs(w http.ResponseWriter, r *http.Request) {
  if !checkUser(w, r) {
    return
  }
  ctx := appengine.NewContext(r)
  q := datastore.NewQuery("Paragraph")
  var paragraphs []Paragraph
  keys, err := q.GetAll(ctx, &paragraphs)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  var results []string
  for i, p := range paragraphs {
    var buf bytes.Buffer
    b := bufio.NewWriter(&buf)
    fmt.Fprintf(b, `"%s":`, keys[i].Encode())
    enc := json.NewEncoder(b)
    if err := enc.Encode(&p); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    b.Flush()
    results = append(results, buf.String())
  }

  w.Header().Set("Content-type", "application/json; charset=utf-8")
  fmt.Fprintf(w, "{%s}\n", strings.Join(results, ",\n"))
}

func getParagraph(w http.ResponseWriter, r *http.Request) {
  if !checkUser(w, r) {
    return
  }
  ctx := appengine.NewContext(r)
  var p Paragraph
  key, err := datastore.DecodeKey(r.FormValue("key"))
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  err = datastore.Get(ctx, key, &p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  w.Header().Set("Content-type", "application/json; charset=utf-8")
  enc := json.NewEncoder(w)
  if err := enc.Encode(&p); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func saveParagraph(w http.ResponseWriter, r *http.Request) {
  if !checkUser(w, r) {
    return
  }
  ctx := appengine.NewContext(r)
  dec := json.NewDecoder(r.Body)
  var p Paragraph
  if err := dec.Decode(&p); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  var err error
  var key *datastore.Key
  keyString := r.FormValue("key")
  if keyString != "" {
    key, err = datastore.DecodeKey(keyString)
  } else {
    key = datastore.NewIncompleteKey(ctx, "Paragraph", nil)
  }
  if err == nil {
    key, err = datastore.Put(ctx, key, &p)
  }
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-type", "text/plain; charset=utf-8")
  fmt.Fprintf(w, "%s", key.Encode())
}

func saveLink(w http.ResponseWriter, r *http.Request) {
  if !checkUser(w, r) {
    return
  }
  ctx := appengine.NewContext(r)
  dec := json.NewDecoder(r.Body)
  var link Link
  if err := dec.Decode(&link); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if (link.OutLink == "" || link.FromKey == "") {
    http.Error(w, "Invalid link, must specify OutLink and FromKey.",
        http.StatusBadRequest)
    return
  }
  var err error
  var linkKey *datastore.Key
  keyString := r.FormValue("key")
  if keyString != "" {
    linkKey, err = datastore.DecodeKey(keyString)
  } else {
    linkKey = datastore.NewIncompleteKey(ctx, "Link", nil)
  }
  if err == nil && link.ToKey == "" {
    pKey := datastore.NewIncompleteKey(ctx, "Paragraph", nil)
    if pKey, err = datastore.Put(ctx, pKey, new(Paragraph)); err == nil {
      link.ToKey = pKey.Encode()
    }
  }

  if err == nil {
    linkKey, err = datastore.Put(ctx, linkKey, &link)
  }
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-type", "text/plain; charset=utf-8")
  fmt.Fprintf(w, "%s\n%s", linkKey.Encode(), link.ToKey)
}

func checkUser(w http.ResponseWriter, r *http.Request) bool {
  ctx := appengine.NewContext(r)
  u := user.Current(ctx)
  if u == nil {
    url, err := user.LoginURL(ctx, r.URL.String())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return false
    }
    w.Header().Set("Location", url)
    w.WriteHeader(http.StatusFound)
    return false
  }
  if !user.IsAdmin(ctx) {
    http.Error(w, "Unauthorized user.", http.StatusUnauthorized)
    return false
  }
  return true
}

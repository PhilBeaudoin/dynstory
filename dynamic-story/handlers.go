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
    "io/ioutil"
    "net/http"

    "appengine"
    "appengine/user"
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
  http.HandleFunc("/save-paragraph/", saveParagraph)
}

func root(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-type", "text/html; charset=utf-8")
  ctx := appengine.NewContext(r)
  u := user.Current(ctx)
  if u == nil {
    url, err := user.LoginURL(ctx, r.URL.String())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Location", url)
    w.WriteHeader(http.StatusFound)
    return
  }

  if !user.IsAdmin(ctx) {
    http.Error(w, "Unauthorized user.", http.StatusUnauthorized)
    return
  }

  w.Write(mainTemplate)
}

func saveParagraph(w http.ResponseWriter, r *http.Request) {
  ctx := appengine.NewContext(r)
  r.ParseForm()
  ctx.Infof("Url: %v", r.URL)
  ctx.Infof("Body: %v", r.Body)
  ctx.Infof("PostForm: %v", r.PostForm)
}


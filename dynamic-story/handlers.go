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
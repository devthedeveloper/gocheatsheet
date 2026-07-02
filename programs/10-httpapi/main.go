// http: a JSON API server AND its client, one file.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Pong struct {
	Msg string `json:"msg"`
	N   int    `json:"n"`
}

func main() {
	mux := http.NewServeMux()
	n := 0
	mux.HandleFunc("GET /ping",
		func(w http.ResponseWriter, r *http.Request) {
			n++
			json.NewEncoder(w).Encode(Pong{"pong", n})
		})

	go http.ListenAndServe(":8090", mux) // serve…
	time.Sleep(100 * time.Millisecond)   // …boot up

	for range 3 { // …and call ourselves
		res, err := http.Get(
			"http://localhost:8090/ping")
		if err != nil {
			panic(err)
		}
		var p Pong
		json.NewDecoder(res.Body).Decode(&p)
		res.Body.Close()
		fmt.Printf("%d ← %+v\n", res.StatusCode, p)
	}
}

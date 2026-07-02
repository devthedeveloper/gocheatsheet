// http: GET /orders/{id} — the service AND its caller.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Order struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /orders/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")
			// (a real handler queries the DB here)
			json.NewEncoder(w).Encode(
				Order{ID: id, Status: "shipped"})
		})

	go http.ListenAndServe(":8090", mux) // the service
	time.Sleep(100 * time.Millisecond)   // let it boot

	// another service asks about order 1042
	res, err := http.Get(
		"http://localhost:8090/orders/1042")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	var o Order
	json.NewDecoder(res.Body).Decode(&o)
	fmt.Println("status code:", res.StatusCode)
	fmt.Printf("order: %+v\n", o)
}

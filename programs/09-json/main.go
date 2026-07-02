// json: the request and response bodies of POST /orders.
package main

import (
	"encoding/json"
	"fmt"
)

type Order struct {
	ID     int      `json:"id"`
	UserID int      `json:"user_id"`
	Items  []string `json:"items"`
	Coupon string   `json:"coupon,omitempty"`
}

func main() {
	// what the client POSTs
	body := `{"user_id": 7, "items": ["boat-airdopes"]}`
	var in Order
	if err := json.Unmarshal([]byte(body), &in); err != nil {
		panic(err) // in a handler: respond 400
	}
	fmt.Printf("decoded request: %+v\n", in)

	// what your handler sends back
	in.ID = 1042 // the DB assigned it
	out, _ := json.MarshalIndent(in, "", "  ")
	fmt.Println(string(out))
}

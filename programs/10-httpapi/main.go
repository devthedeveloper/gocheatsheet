// http: GET /orders/{id} — a real handler querying real MySQL,
// called by a real HTTP client.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Order struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

func main() {
	db, err := sql.Open("mysql",
		"root:secret@tcp(localhost:3306)/notes")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.Exec("DROP TABLE IF EXISTS http_orders")
	db.Exec(`CREATE TABLE http_orders (
		id INT PRIMARY KEY, status VARCHAR(20))`)
	db.Exec("INSERT INTO http_orders VALUES (1042, 'shipped')")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /orders/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")
			var o Order
			err := db.QueryRow(
				"SELECT id, status FROM http_orders WHERE id=?",
				id).Scan(&o.ID, &o.Status)
			if err == sql.ErrNoRows {
				http.Error(w, "not found", 404)
				return
			}
			json.NewEncoder(w).Encode(o) // REAL row → REAL JSON
		})

	go http.ListenAndServe(":8090", mux) // the service
	time.Sleep(100 * time.Millisecond)   // let it bind the port

	// another service asks about order 1042 — a REAL HTTP round trip
	res, err := http.Get("http://localhost:8090/orders/1042")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	var o Order
	json.NewDecoder(res.Body).Decode(&o)
	fmt.Println("status code:", res.StatusCode)
	fmt.Printf("order: %+v\n", o)
}

// json: decode a real request body, save it, marshal the real DB row back.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Order struct {
	ID     int    `json:"id"`
	UserID int    `json:"user_id"`
	Item   string `json:"item"`
	Coupon string `json:"coupon,omitempty"`
}

func main() {
	db, err := sql.Open("mysql",
		"root:secret@tcp(localhost:3306)/notes")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.Exec("DROP TABLE IF EXISTS json_orders")
	db.Exec(`CREATE TABLE json_orders (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT, item VARCHAR(80))`)

	// exactly what a client POSTs to /orders
	body := `{"user_id": 7, "item": "boat-airdopes"}`
	var in Order
	if err := json.Unmarshal([]byte(body), &in); err != nil {
		panic(err) // in a handler: respond 400
	}
	fmt.Printf("decoded request: %+v\n", in)

	// REAL insert — the DB assigns the id, not the client
	res, err := db.Exec(
		"INSERT INTO json_orders (user_id, item) VALUES (?, ?)",
		in.UserID, in.Item)
	if err != nil {
		panic(err)
	}
	id, _ := res.LastInsertId()

	// REAL read-back — the response reflects what's actually stored
	var out Order
	db.QueryRow("SELECT id, user_id, item FROM json_orders WHERE id=?", id).
		Scan(&out.ID, &out.UserID, &out.Item)

	body2, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(body2))
}

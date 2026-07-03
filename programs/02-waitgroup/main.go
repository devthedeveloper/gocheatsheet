// wg.Wait: GET /dashboard runs 3 real MySQL queries concurrently,
// then answers only once ALL three have returned.
package main

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

func setup(db *sql.DB) {
	db.Exec("DROP TABLE IF EXISTS orders, users")
	db.Exec(`CREATE TABLE users (id INT PRIMARY KEY)`)
	db.Exec(`CREATE TABLE orders (
		id INT PRIMARY KEY, total_paisa INT)`)
	for i := 1; i <= 500; i++ {
		db.Exec("INSERT INTO users VALUES (?)", i)
	}
	for i := 1; i <= 1200; i++ {
		db.Exec("INSERT INTO orders VALUES (?, ?)",
			i, 19900+(i%5)*500)
	}
}

func main() {
	db, err := sql.Open("mysql",
		"root:secret@tcp(localhost:3306)/notes")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	setup(db) // seed data, so this file is rerunnable standalone

	var (
		wg                          sync.WaitGroup
		userCount, orderCount       int
		revenuePaisa                int64
	)
	fmt.Println("GET /dashboard → 3 real queries, in parallel…")

	wg.Add(1) // one more query in flight
	go func() {
		defer wg.Done() // crossed off, always
		db.QueryRow("SELECT COUNT(*) FROM users").
			Scan(&userCount)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		db.QueryRow("SELECT COUNT(*) FROM orders").
			Scan(&orderCount)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		db.QueryRow("SELECT COALESCE(SUM(total_paisa),0) " +
			"FROM orders").Scan(&revenuePaisa)
	}()

	wg.Wait() // block here until all three queries return

	fmt.Printf("users:   %d\n", userCount)
	fmt.Printf("orders:  %d\n", orderCount)
	fmt.Printf("revenue: ₹%.2f\n", float64(revenuePaisa)/100)
	fmt.Println("200 OK — response assembled ✓")
}

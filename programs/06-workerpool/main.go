// worker pool: 9 pending orders, only 3 DB connections fulfilling them.
package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func setup(db *sql.DB) {
	db.Exec("DROP TABLE IF EXISTS fulfillment")
	db.Exec(`CREATE TABLE fulfillment (
		id INT PRIMARY KEY, status VARCHAR(20))`)
	for i := 1; i <= 9; i++ {
		db.Exec("INSERT INTO fulfillment VALUES (?, 'pending')", i)
	}
}

func main() {
	db, err := sql.Open("mysql",
		"root:secret@tcp(localhost:3306)/notes")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(3) // the pool really can't exceed 3 connections
	setup(db)

	start := time.Now()
	orderIDs := make(chan int)
	var wg sync.WaitGroup

	for conn := 1; conn <= 3; conn++ { // 3 real DB connections
		wg.Add(1)
		go func(conn int) {
			defer wg.Done()
			for id := range orderIDs { // until closed
				// a REAL write, through a REAL pooled connection
				_, err := db.Exec(
					"UPDATE fulfillment SET status='shipped' WHERE id=?",
					id)
				if err != nil {
					fmt.Println("update failed:", err)
					continue
				}
				fmt.Printf("conn %d shipped order #%d\n", conn, id)
			}
		}(conn)
	}

	for i := 1; i <= 9; i++ {
		orderIDs <- i
	}
	close(orderIDs)
	wg.Wait()

	var shipped int
	db.QueryRow("SELECT COUNT(*) FROM fulfillment WHERE status='shipped'").
		Scan(&shipped)
	fmt.Printf("%d/9 orders shipped, in %v\n",
		shipped, time.Since(start).Round(time.Millisecond))
}

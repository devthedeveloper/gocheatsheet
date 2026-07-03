// context: a slow report query, cancelled by the request's real deadline.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql",
		"root:secret@tcp(localhost:3306)/notes")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// this endpoint has a 400ms budget
	ctx, cancel := context.WithTimeout(
		context.Background(), 400*time.Millisecond)
	defer cancel() // always, even on success

	fmt.Println("generating report (query takes 5s)…")
	start := time.Now()

	// SELECT SLEEP(5) is a REAL, genuinely slow query — and
	// QueryContext really cancels it on the wire when ctx expires
	var dummy int
	err = db.QueryRowContext(ctx, "SELECT SLEEP(5)").Scan(&dummy)

	fmt.Println("took:", time.Since(start).Round(time.Millisecond))
	if err != nil {
		fmt.Println("aborting report:", err)
		fmt.Println("handler: told client 504, moved on")
		return
	}
	fmt.Println("report ready (this line never runs)")
}

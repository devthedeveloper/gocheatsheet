// select: an in-process cache vs a genuinely slow MySQL query —
// serve whichever answers first, inside a real time budget.
package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// a real (tiny) in-memory cache — no network, just a map
var cartCache = map[string]string{
	"cart:42": "3 items (cache)",
}

func cacheGet(key string) <-chan string {
	ch := make(chan string, 1)
	go func() {
		if v, ok := cartCache[key]; ok {
			ch <- v // real map lookup, no artificial delay needed
		}
	}()
	return ch
}

func dbQuery(db *sql.DB, key string) <-chan string {
	ch := make(chan string, 1)
	go func() {
		var dummy int
		// SLEEP(3) runs INSIDE MySQL — a genuinely slow query,
		// not a Go-side time.Sleep pretending to be one
		err := db.QueryRow("SELECT SLEEP(3)").Scan(&dummy)
		if err == nil {
			ch <- key + " → 3 items (db, after a real 3s query)"
		}
	}()
	return ch
}

func main() {
	db, err := sql.Open("mysql",
		"root:secret@tcp(localhost:3306)/notes")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	cache := cacheGet("cart:42")
	dbCh := dbQuery(db, "cart:42")

	select {
	case v := <-cache:
		fmt.Println("hit:", v)
	case v := <-dbCh:
		fmt.Println("hit:", v)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("both slow → 504, don't hang")
	}
	fmt.Println("responded within the 100ms budget ✓")
}

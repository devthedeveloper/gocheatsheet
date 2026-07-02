// select: cache vs DB — serve the fastest, respect the budget.
package main

import (
	"fmt"
	"time"
)

func cacheGet(key string) <-chan string {
	ch := make(chan string, 1)
	go func() {
		time.Sleep(5 * time.Millisecond) // redis is quick
		ch <- key + " → 3 items (cache)"
	}()
	return ch
}

func dbQuery(key string) <-chan string {
	ch := make(chan string, 1)
	go func() {
		time.Sleep(3 * time.Second) // db is having a day
		ch <- key + " → 3 items (db)"
	}()
	return ch
}

func main() {
	cache := cacheGet("cart:42")
	db := dbQuery("cart:42")

	select {
	case v := <-cache:
		fmt.Println("hit:", v)
	case v := <-db:
		fmt.Println("hit:", v)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("both slow → 504, don't hang")
	}
	fmt.Println("responded within the 100ms budget ✓")
}

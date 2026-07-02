// select: wait on several channels — first ready wins.
package main

import (
	"fmt"
	"time"
)

func worker(name string, d time.Duration) <-chan string {
	ch := make(chan string, 1)
	go func() {
		time.Sleep(d)
		ch <- name + " finished"
	}()
	return ch
}

func main() {
	fast := worker("fast", 50*time.Millisecond)
	slow := worker("slow", 5*time.Second)

	for {
		select {
		case msg := <-fast:
			fmt.Println(msg)
		case msg := <-slow:
			fmt.Println(msg) // never happens
		case <-time.After(200 * time.Millisecond):
			fmt.Println("⏰ timeout — not waiting 5s for slow")
			return
		}
	}
}

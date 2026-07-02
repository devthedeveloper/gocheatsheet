// worker pool: 9 newsletters, only 3 SMTP connections.
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	start := time.Now()
	queue := make(chan string)
	var wg sync.WaitGroup

	for conn := 1; conn <= 3; conn++ { // 3 connections
		wg.Add(1)
		go func() {
			defer wg.Done()
			for email := range queue { // until closed
				time.Sleep(100 * time.Millisecond) // send
				fmt.Printf("conn %d sent → %s\n",
					conn, email)
			}
		}()
	}

	users := []string{
		"asha", "ravi", "meera", "arjun", "divya",
		"karan", "nisha", "rahul", "sneha",
	}
	for _, u := range users {
		queue <- u + "@example.com"
	}
	close(queue) // newsletter fully queued
	wg.Wait()

	fmt.Printf("9 emails on 3 connections = %v\n",
		time.Since(start).Round(10*time.Millisecond))
}

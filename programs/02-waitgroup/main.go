// wg.Wait: GET /dashboard answers only after ALL services do.
package main

import (
	"fmt"
	"sync"
	"time"
)

func call(service string, ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
	fmt.Printf("%s answered in %dms\n", service, ms)
}

func main() {
	var wg sync.WaitGroup

	// the dashboard needs three services
	services := map[string]int{
		"profile-service":       100,
		"orders-service":        250,
		"notifications-service": 180,
	}
	fmt.Println("GET /dashboard → calling 3 services…")
	for name, latency := range services {
		wg.Add(1) // one more call in flight
		go func() {
			defer wg.Done() // crossed off, always
			call(name, latency)
		}()
	}

	wg.Wait() // block here until all three answer
	fmt.Println("200 OK — response assembled ✓")
}

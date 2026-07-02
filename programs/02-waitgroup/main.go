// waitgroup: wait for N goroutines to finish.
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	for id := 1; id <= 3; id++ {
		wg.Add(1) // +1 BEFORE starting
		go func() {
			defer wg.Done() // -1, always
			wait := time.Duration(id) *
				100 * time.Millisecond
			time.Sleep(wait) // fake work
			fmt.Printf("worker %d done (%v)\n",
				id, wait)
		}()
	}

	fmt.Println("waiting for 3 workers…")
	wg.Wait() // blocks until counter hits 0
	fmt.Println("all done ✓")
}

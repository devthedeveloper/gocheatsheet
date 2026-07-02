// mutex: what a data race costs, and the fix.
package main

import (
	"fmt"
	"sync"
)

func main() {
	const workers, bumps = 100, 1000
	var wg sync.WaitGroup

	// BROKEN: 100 goroutines bump a plain int
	unsafe := 0
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range bumps {
				unsafe++ // RACE: read+write
			}
		}()
	}
	wg.Wait()

	// FIXED: same thing behind a mutex
	var mu sync.Mutex
	safe := 0
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range bumps {
				mu.Lock()
				safe++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	fmt.Println("expected:      ", workers*bumps)
	fmt.Println("without mutex: ", unsafe, "😱")
	fmt.Println("with mutex:    ", safe, "✓")
	// try: go run -race . → the race is caught
}

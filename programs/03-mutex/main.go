// mutex: 100 concurrent clients earn ₹1 cashback per hit.
package main

import (
	"fmt"
	"sync"
)

func main() {
	const clients, hits = 100, 1000
	var wg sync.WaitGroup

	// BROKEN: every request credits the wallet, no lock
	balance := 0
	for range clients {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range hits {
				balance++ // RACE: read, add, write
			}
		}()
	}
	wg.Wait()

	// FIXED: same traffic, one mutex
	var mu sync.Mutex
	safe := 0
	for range clients {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range hits {
				mu.Lock()
				safe++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	fmt.Println("should be: ₹", clients*hits)
	fmt.Println("no mutex:  ₹", balance, "— money vanished 😱")
	fmt.Println("mutex:     ₹", safe, "✓")
}

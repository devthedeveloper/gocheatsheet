// worker pool: N goroutines chew through a job queue.
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	start := time.Now()
	jobs := make(chan int)
	results := make(chan string)
	var wg sync.WaitGroup

	for w := 1; w <= 3; w++ { // exactly 3 workers
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs { // until closed
				time.Sleep(100 * time.Millisecond)
				results <- fmt.Sprintf(
					"worker %d finished job %d", w, j)
			}
		}()
	}

	go func() { // feed 9 jobs, then close
		for j := 1; j <= 9; j++ {
			jobs <- j
		}
		close(jobs)
	}()

	go func() { // close results when workers exit
		wg.Wait()
		close(results)
	}()

	for r := range results {
		fmt.Println(r)
	}
	fmt.Printf("9 jobs × 100ms on 3 workers = %v\n",
		time.Since(start).Round(10*time.Millisecond))
}

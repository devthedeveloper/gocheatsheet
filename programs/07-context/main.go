// context: a report that stops when the request times out.
package main

import (
	"context"
	"fmt"
	"time"
)

func generateReport(ctx context.Context) {
	for batch := 1; ; batch++ {
		select {
		case <-ctx.Done(): // out of time / client gone
			fmt.Println("aborting report:", ctx.Err())
			return
		case <-time.After(100 * time.Millisecond):
			fmt.Printf("processed batch %d (10k rows)\n",
				batch)
		}
	}
}

func main() {
	// this endpoint has a 350ms budget
	ctx, cancel := context.WithTimeout(
		context.Background(), 350*time.Millisecond)
	defer cancel()

	generateReport(ctx)
	fmt.Println("handler: told client 504, moved on")
}

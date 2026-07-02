// context: the universal "please stop" signal.
package main

import (
	"context"
	"fmt"
	"time"
)

func tick(ctx context.Context) {
	for i := 1; ; i++ {
		select {
		case <-ctx.Done(): // cancelled or timed out
			fmt.Println("worker: stopping —", ctx.Err())
			return
		case <-time.After(100 * time.Millisecond):
			fmt.Println("worker: tick", i)
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(), 350*time.Millisecond)
	defer cancel() // always, even on success

	tick(ctx) // returns when the context fires
	fmt.Println("main: clean exit")
}

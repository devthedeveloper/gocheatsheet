// channels: typed pipes between goroutines.
package main

import "fmt"

func main() {
	ch := make(chan string)

	go func() { // the sender
		menu := []string{"🍎", "🍌", "🍇"}
		for _, fruit := range menu {
			fmt.Println("sending", fruit)
			ch <- fruit // blocks until received
		}
		close(ch) // sender says: no more
	}()

	for f := range ch { // receive until closed
		fmt.Println("got    ", f)
	}
	fmt.Println("channel drained, bye")
}

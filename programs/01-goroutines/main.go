// goroutines: run things concurrently with one keyword.
package main

import (
	"fmt"
	"time"
)

func say(name string) {
	for i := 1; i <= 3; i++ {
		fmt.Println(name, i)
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	go say("A") // fires and returns instantly
	go say("B")

	say("main") // runs in the main goroutine,
	//             keeping the program alive

	// when main returns, ALL goroutines die.
	// delete the say("main") line and watch
	// A and B never get a chance to print.
}

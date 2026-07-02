// goroutines: three signup emails, sent at the same time.
package main

import (
	"fmt"
	"time"
)

func sendEmail(to string) {
	fmt.Println("connecting to SMTP for", to)
	time.Sleep(150 * time.Millisecond) // SMTP round-trip
	fmt.Println("✉️  delivered to", to)
}

func main() {
	// three users just signed up
	go sendEmail("asha@example.com")
	go sendEmail("ravi@example.com")

	sendEmail("meera@example.com") // main sends one too —
	// which keeps the program alive long enough

	// when main returns, ALL goroutines die.
	// make main's send a `go` too and the program
	// exits instantly: zero emails delivered.
}

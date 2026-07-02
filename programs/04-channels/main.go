// channels: uploaded files handed to the virus scanner.
package main

import "fmt"

func main() {
	uploads := make(chan string) // the hand-off point

	go func() { // goroutine A: the upload receiver
		files := []string{
			"invoice_jan.pdf",
			"invoice_feb.pdf",
			"salary_slip.pdf",
		}
		for _, f := range files {
			fmt.Println("upload complete:", f)
			uploads <- f // hand to the scanner
		}
		close(uploads) // no more uploads today
	}()

	for f := range uploads { // main: the scanner
		fmt.Println("  virus-scanned:", f)
	}
	fmt.Println("all uploads processed ✓")
}

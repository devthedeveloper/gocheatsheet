// errgroup: fan out, collect the first error.
// needs: go get golang.org/x/sync/errgroup
package main

import (
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"
)

func fetch(url string) error {
	if url == "https://bad.example" {
		return errors.New("boom: " + url)
	}
	fmt.Println("fetched", url)
	return nil
}

func main() {
	var g errgroup.Group
	g.SetLimit(2) // at most 2 at a time

	urls := []string{
		"https://go.dev",
		"https://bad.example",
		"https://github.com",
	}
	for _, u := range urls {
		g.Go(func() error {
			return fetch(u)
		})
	}

	// Wait returns the FIRST error (or nil)
	if err := g.Wait(); err != nil {
		fmt.Println("g.Wait() →", err)
	}
}

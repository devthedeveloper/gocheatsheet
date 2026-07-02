// errgroup: POST /checkout fans out; payment declines.
// needs: go get golang.org/x/sync/errgroup
package main

import (
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"
)

func callService(name string) error {
	if name == "payment-service" {
		return errors.New(
			"payment-service: card declined")
	}
	fmt.Println(name, "→ OK")
	return nil
}

func main() {
	var g errgroup.Group

	for _, svc := range []string{
		"inventory-service",
		"payment-service",
		"shipping-service",
	} {
		g.Go(func() error {
			return callService(svc)
		})
	}

	// Wait = wg.Wait() + the first error
	if err := g.Wait(); err != nil {
		fmt.Println("checkout failed:", err)
		fmt.Println("→ respond 402, release the stock")
	}
}

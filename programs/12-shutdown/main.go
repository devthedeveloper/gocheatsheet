// graceful shutdown: deploy day for your orders API.
// (sends ITSELF a SIGTERM so you can watch the drain)
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	srv := &http.Server{Addr: ":8091"}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	go srv.ListenAndServe()
	fmt.Println("orders API serving on :8091")

	go func() { // pretend k8s deploys in 300ms
		time.Sleep(300 * time.Millisecond)
		syscall.Kill(os.Getpid(), syscall.SIGTERM)
	}()

	<-ctx.Done() // block until the signal
	fmt.Println("SIGTERM: stop taking orders, finish the open ones…")

	shCtx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		panic(err)
	}
	fmt.Println("all in-flight requests served — clean exit ✓")
}

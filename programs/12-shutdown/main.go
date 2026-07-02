// graceful shutdown: drain requests before dying.
// (sends ITSELF a SIGTERM to simulate a deploy)
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
	fmt.Println("serving on :8091")

	go func() { // pretend k8s kills us in 300ms
		time.Sleep(300 * time.Millisecond)
		syscall.Kill(os.Getpid(), syscall.SIGTERM)
	}()

	<-ctx.Done() // block until the signal
	fmt.Println("SIGTERM received, draining…")

	shCtx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		panic(err)
	}
	fmt.Println("in-flight requests done, clean exit ✓")
}

// gopherit — a Reddit-clone REST API built with Go's standard library.
// This is the companion project for the "Go + REST: Build Your Own Reddit"
// handwritten-notes ebook.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopherit/internal/store"
)

const version = "1.0.0"

// config holds everything tweakable from the outside world.
type config struct {
	port int
	env  string
	dsn  string
}

// application is the "toolbox" every handler can reach: config, a logger
// and the data layer. Handlers hang off it as methods.
type application struct {
	config  config
	logger  *slog.Logger
	store   *store.Store
	feed    *feedCache
	limiter *ipLimiter
}

func main() {
	var cfg config
	var maxConns int
	var rlRPS float64
	var rlBurst int
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|production)")
	flag.StringVar(&cfg.dsn, "dsn", "gopherit.db", "SQLite database file")
	flag.IntVar(&maxConns, "maxconns", 1, "max open DB connections (see chapter 4)")
	flag.Float64Var(&rlRPS, "rate", 0, "per-IP requests/sec (0 = off; see chapter 7)")
	flag.IntVar(&rlBurst, "burst", 20, "per-IP burst size")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	st, err := store.Open(cfg.dsn)
	if err != nil {
		logger.Error("cannot open database", "error", err)
		os.Exit(1)
	}
	defer st.Close()
	st.SetMaxOpenConns(maxConns)
	logger.Info("database ready", "dsn", cfg.dsn, "maxconns", maxConns)

	// Recompute post hotness in the background so the feed query never pays for
	// pow() on the request path (chapter 4).
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := st.RefreshHotness(); err != nil {
				logger.Error("hotness refresh failed", "error", err)
			}
		}
	}()

	app := &application{
		config: cfg,
		logger: logger,
		store:  st,
		feed:   newFeedCache(time.Second), // the front page is the same for everyone
	}
	if rlRPS > 0 {
		app.limiter = newIPLimiter(rlRPS, rlBurst)
	}

	// Profiling endpoints (chapter 3), on a SEPARATE localhost-only port so
	// they're never exposed to the public internet. Visit
	// http://localhost:6060/debug/pprof/ while the server is under load.
	if cfg.env != "production" {
		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
			logger.Info("pprof on http://localhost:6060/debug/pprof/")
			http.ListenAndServe("localhost:6060", mux)
		}()
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown: when Ctrl+C (SIGINT) or SIGTERM arrives, stop
	// accepting new requests and give in-flight ones 10s to finish.
	shutdownErr := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit

		logger.Info("shutting down", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdownErr <- srv.Shutdown(ctx)
	}()

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env, "version", version)

	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
	if err := <-shutdownErr; err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped gracefully")
}

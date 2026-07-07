package main

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// rateLimit is a per-IP token-bucket limiter. bcrypt makes POST /tokens
// expensive on purpose, which turns it into a denial-of-service target: a few
// dozen concurrent logins pin every CPU. This middleware caps how fast any one
// IP can hit us, rejecting the flood with a cheap 429 *before* it reaches
// bcrypt — so real traffic keeps flowing.
type ipLimiter struct {
	mu      sync.Mutex
	clients map[string]*client
	rps     rate.Limit
	burst   int
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newIPLimiter(rps float64, burst int) *ipLimiter {
	l := &ipLimiter{clients: make(map[string]*client), rps: rate.Limit(rps), burst: burst}
	// A janitor goroutine evicts IPs we haven't seen in a while, so the map
	// can't grow without bound.
	go func() {
		for {
			time.Sleep(time.Minute)
			l.mu.Lock()
			for ip, c := range l.clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(l.clients, ip)
				}
			}
			l.mu.Unlock()
		}
	}()
	return l
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	c, ok := l.clients[ip]
	if !ok {
		c = &client{limiter: rate.NewLimiter(l.rps, l.burst)}
		l.clients[ip] = c
	}
	c.lastSeen = time.Now()
	return c.limiter.Allow()
}

// rateLimit wraps a handler, rejecting over-limit IPs with 429. If no limiter
// is configured (the -rate flag is 0) it does nothing.
func (app *application) rateLimit(next http.Handler) http.Handler {
	if app.limiter == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if !app.limiter.allow(ip) {
			w.Header().Set("Retry-After", "1")
			app.errorResponse(w, http.StatusTooManyRequests, "rate limit exceeded — slow down")
			return
		}
		next.ServeHTTP(w, r)
	})
}

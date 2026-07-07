// hammer is a tiny HTTP load tester we build in chapter 2. It fires requests
// from a pool of goroutine workers, records every latency, and prints the
// numbers that actually matter: throughput, error rate, and p50/p95/p99.
//
// It is deliberately ~150 lines of standard library — the point is to
// understand every number it prints. For serious work, reach for k6/vegeta/wrk.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type result struct {
	total   int64
	errors  int64
	mu      sync.Mutex
	latency []time.Duration // guarded by mu
}

func (r *result) record(d time.Duration, ok bool) {
	atomic.AddInt64(&r.total, 1)
	if !ok {
		atomic.AddInt64(&r.errors, 1)
	}
	r.mu.Lock()
	r.latency = append(r.latency, d)
	r.mu.Unlock()
}

func main() {
	var (
		base     = flag.String("url", "http://localhost:4000", "base URL")
		workers  = flag.Int("workers", 50, "concurrent workers")
		duration = flag.Duration("duration", 15*time.Second, "how long to run")
		scenario = flag.String("scenario", "feed", "feed | mixed | login")
		nTokens  = flag.Int("tokens", 20, "users to log in for authed scenarios")
		nPosts   = flag.Int("posts", 50000, "highest seeded post id")
		nUsers   = flag.Int("users", 10000, "highest seeded user id")
	)
	flag.Parse()

	client := &http.Client{Timeout: 30 * time.Second}

	// Preflight: grab a pool of tokens for scenarios that write.
	var tokens []string
	if *scenario == "mixed" {
		tokens = login(client, *base, *nTokens, *nUsers)
		if len(tokens) == 0 {
			log.Fatal("mixed scenario needs tokens but all logins failed")
		}
		log.Printf("logged in %d users", len(tokens))
	}

	res := &result{}
	deadline := time.Now().Add(*duration)
	var wg sync.WaitGroup
	log.Printf("firing: scenario=%s workers=%d duration=%s", *scenario, *workers, *duration)
	realStart := time.Now()

	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func(seed int64) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(seed))
			for time.Now().Before(deadline) {
				start := time.Now()
				ok := doRequest(client, *base, *scenario, rng, tokens, *nPosts)
				res.record(time.Since(start), ok)
			}
		}(int64(w) + 1)
	}
	wg.Wait()
	report(res, time.Since(realStart))
}

// doRequest performs one request chosen by the scenario's weights.
func doRequest(c *http.Client, base, scenario string, rng *rand.Rand, tokens []string, nPosts int) bool {
	switch scenario {
	case "login":
		return post(c, base+"/api/v1/tokens", "",
			map[string]any{"email": fmt.Sprintf("user%d@example.com", rng.Intn(10000)+1), "password": "password123"})
	case "mixed":
		n := rng.Intn(100)
		switch {
		case n < 90: // 90% reads
			return get(c, fmt.Sprintf("%s/api/v1/posts?sort=hot&page=%d", base, rng.Intn(20)+1))
		case n < 97: // 7% votes
			tok := tokens[rng.Intn(len(tokens))]
			return post(c, fmt.Sprintf("%s/api/v1/posts/%d/vote", base, rng.Intn(nPosts)+1), tok,
				map[string]any{"value": 1})
		default: // 3% new posts
			tok := tokens[rng.Intn(len(tokens))]
			return post(c, base+"/api/v1/posts", tok,
				map[string]any{"subreddit": fmt.Sprintf("sub%d", rng.Intn(200)+1),
					"title": "load test post", "body": "hello"})
		}
	default: // "feed"
		return get(c, fmt.Sprintf("%s/api/v1/posts?sort=hot&page=%d", base, rng.Intn(20)+1))
	}
}

func login(c *http.Client, base string, n, nUsers int) []string {
	var tokens []string
	for i := 1; i <= n; i++ {
		body, _ := json.Marshal(map[string]any{
			"email": fmt.Sprintf("user%d@example.com", i), "password": "password123"})
		resp, err := c.Post(base+"/api/v1/tokens", "application/json", bytes.NewReader(body))
		if err != nil {
			continue
		}
		var out struct {
			Token string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&out)
		resp.Body.Close()
		if out.Token != "" {
			tokens = append(tokens, out.Token)
		}
	}
	return tokens
}

func get(c *http.Client, url string) bool {
	resp, err := c.Get(url)
	if err != nil {
		return false
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode < 400
}

func post(c *http.Client, url, token string, body map[string]any) bool {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.Do(req)
	if err != nil {
		return false
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode < 400
}

func report(r *result, elapsed time.Duration) {
	r.mu.Lock()
	lat := r.latency
	r.mu.Unlock()
	sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })

	pct := func(p float64) time.Duration {
		if len(lat) == 0 {
			return 0
		}
		i := int(p / 100 * float64(len(lat)))
		if i >= len(lat) {
			i = len(lat) - 1
		}
		return lat[i]
	}

	rps := float64(r.total) / elapsed.Seconds()
	fmt.Printf("\n─────────── results ───────────\n")
	fmt.Printf("requests   : %d in %s\n", r.total, elapsed.Round(time.Millisecond))
	fmt.Printf("throughput : %.0f req/s\n", rps)
	fmt.Printf("errors     : %d (%.2f%%)\n", r.errors, 100*float64(r.errors)/float64(max(r.total, 1)))
	fmt.Printf("latency p50: %s\n", pct(50).Round(time.Millisecond))
	fmt.Printf("latency p95: %s\n", pct(95).Round(time.Millisecond))
	fmt.Printf("latency p99: %s\n", pct(99).Round(time.Millisecond))
	fmt.Printf("latency max: %s\n", pct(100).Round(time.Millisecond))
	fmt.Printf("───────────────────────────────\n")
}

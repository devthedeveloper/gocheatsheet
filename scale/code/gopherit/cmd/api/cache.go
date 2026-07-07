package main

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// feedCache is a tiny in-process cache for the front page. The feed is the
// same bytes for everyone, so there is no reason to rebuild it from the
// database on every request — we build it once and hand the same bytes to
// everyone who asks within the TTL.
//
// singleflight solves the "thundering herd": when a cached entry expires under
// heavy traffic, only ONE goroutine rebuilds it; the rest wait for that single
// result instead of all hammering the database at once.
type feedCache struct {
	ttl   time.Duration
	group singleflight.Group

	mu      sync.RWMutex
	entries map[string]cacheEntry
}

type cacheEntry struct {
	data    []byte
	expires time.Time
}

func newFeedCache(ttl time.Duration) *feedCache {
	return &feedCache{ttl: ttl, entries: make(map[string]cacheEntry)}
}

// get returns cached bytes for key if they are still fresh.
func (c *feedCache) get(key string) ([]byte, bool) {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expires) {
		return nil, false
	}
	return e.data, true
}

// fetch returns cached bytes, or builds them via build() if missing/stale.
// Concurrent callers for the same key share a single build.
func (c *feedCache) fetch(key string, build func() ([]byte, error)) ([]byte, error) {
	if data, ok := c.get(key); ok {
		return data, nil
	}
	data, err, _ := c.group.Do(key, func() (any, error) {
		// Re-check: another goroutine may have filled it while we queued.
		if data, ok := c.get(key); ok {
			return data, nil
		}
		fresh, err := build()
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		c.entries[key] = cacheEntry{data: fresh, expires: time.Now().Add(c.ttl)}
		c.mu.Unlock()
		return fresh, nil
	})
	if err != nil {
		return nil, err
	}
	return data.([]byte), nil
}

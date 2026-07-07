// linkboard — a tiny anonymous link board, chapter 1: everything in one file,
// with the links kept in memory. No database yet (that's chapter 2).
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

// Link is one submission on the board. The `json:"..."` tags decide the key
// names in the JSON we send back.
type Link struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Votes int64  `json:"votes"`
}

// Our "database" for now: a slice in memory. A mutex guards it because the
// server runs each request in its own goroutine, so two POSTs can arrive at
// the same instant. (Goroutines & mutexes: see the cheatsheet.)
var (
	mu     sync.Mutex
	links  []Link
	nextID int64 = 1
)

// listLinksHandler — GET /links
func listLinksHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

// createLinkHandler — POST /links   Body: {"title": "...", "url": "..."}
func createLinkHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "body must be valid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	link := Link{ID: nextID, Title: input.Title, URL: input.URL}
	nextID++
	links = append(links, link)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(link)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /links", listLinksHandler)
	mux.HandleFunc("POST /links", createLinkHandler)

	log.Println("linkboard listening on http://localhost:4000")
	log.Fatal(http.ListenAndServe(":4000", mux))
}

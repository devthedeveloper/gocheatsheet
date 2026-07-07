// linkboard, chapter 3: upvotes and sorting. New endpoint POST /links/{id}/vote,
// and GET /links now understands ?sort=top or ?sort=new.
package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

// listLinksHandler — GET /links?sort=new|top
func listLinksHandler(w http.ResponseWriter, r *http.Request) {
	sort := r.URL.Query().Get("sort") // "" when absent -> newest first
	links, err := listLinks(sort)
	if err != nil {
		http.Error(w, "could not load links", http.StatusInternalServerError)
		return
	}
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

	link, err := insertLink(input.Title, input.URL)
	if err != nil {
		http.Error(w, "could not save link", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(link)
}

// voteHandler — POST /links/{id}/vote
func voteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "no such link", http.StatusNotFound)
		return
	}

	votes, err := voteLink(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "no such link", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "could not vote", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"id": id, "votes": votes})
}

func main() {
	if err := openDB("linkboard.db"); err != nil {
		log.Fatal(err)
	}
	log.Println("database ready: linkboard.db")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /links", listLinksHandler)
	mux.HandleFunc("POST /links", createLinkHandler)
	mux.HandleFunc("POST /links/{id}/vote", voteHandler)

	log.Println("linkboard listening on http://localhost:4000")
	log.Fatal(http.ListenAndServe(":4000", mux))
}

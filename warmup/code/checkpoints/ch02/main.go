// linkboard, chapter 2: the links now live in a SQLite file, so they survive a
// restart. The in-memory slice and its mutex are gone — the database handles
// concurrent access for us.
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// listLinksHandler — GET /links
func listLinksHandler(w http.ResponseWriter, r *http.Request) {
	links, err := listLinks()
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

func main() {
	if err := openDB("linkboard.db"); err != nil {
		log.Fatal(err)
	}
	log.Println("database ready: linkboard.db")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /links", listLinksHandler)
	mux.HandleFunc("POST /links", createLinkHandler)

	log.Println("linkboard listening on http://localhost:4000")
	log.Fatal(http.ListenAndServe(":4000", mux))
}

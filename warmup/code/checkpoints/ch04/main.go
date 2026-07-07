// linkboard, chapter 4: a friendlier API (JSON errors + validation) and a tiny
// web page served right from the same server. main.go is now just wiring.
package main

import (
	"log"
	"net/http"
)

func main() {
	if err := openDB("linkboard.db"); err != nil {
		log.Fatal(err)
	}
	log.Println("database ready: linkboard.db")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", homeHandler) // the web page, only at "/"
	mux.HandleFunc("GET /links", listLinksHandler)
	mux.HandleFunc("POST /links", createLinkHandler)
	mux.HandleFunc("POST /links/{id}/vote", voteHandler)

	log.Println("linkboard listening on http://localhost:4000")
	log.Fatal(http.ListenAndServe(":4000", mux))
}

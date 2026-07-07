package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// homeHandler serves the single-page frontend at "/".
func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// listLinksHandler — GET /links?sort=new|top
func listLinksHandler(w http.ResponseWriter, r *http.Request) {
	links, err := listLinks(r.URL.Query().Get("sort"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load links")
		return
	}
	writeJSON(w, http.StatusOK, links)
}

// createLinkHandler — POST /links   Body: {"title": "...", "url": "..."}
func createLinkHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "body must be valid JSON")
		return
	}

	// Validate: a title, and a url that actually looks like one.
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" || len(input.Title) > 200 {
		writeError(w, http.StatusUnprocessableEntity, "title must be 1-200 characters")
		return
	}
	if !strings.HasPrefix(input.URL, "http://") && !strings.HasPrefix(input.URL, "https://") {
		writeError(w, http.StatusUnprocessableEntity, "url must start with http:// or https://")
		return
	}

	link, err := insertLink(input.Title, input.URL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not save link")
		return
	}
	writeJSON(w, http.StatusCreated, link)
}

// voteHandler — POST /links/{id}/vote
func voteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "no such link")
		return
	}

	votes, err := voteLink(id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "no such link")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not vote")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"id": id, "votes": votes})
}

// writeJSON sends v as JSON with the given status code. The CORS header lets a
// frontend hosted somewhere else call this API from the browser.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError sends {"error": "..."} with the given status code.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

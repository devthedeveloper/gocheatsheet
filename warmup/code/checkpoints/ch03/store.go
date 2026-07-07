package main

import (
	"database/sql"

	_ "modernc.org/sqlite" // registers the "sqlite" driver with database/sql
)

// Link is one submission on the board.
type Link struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Votes     int64  `json:"votes"`
	CreatedAt string `json:"created_at"`
}

// db is the one shared connection pool.
var db *sql.DB

// openDB connects to the SQLite file and creates the links table if needed.
func openDB(path string) error {
	var err error
	db, err = sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)")
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS links (
            id         INTEGER PRIMARY KEY AUTOINCREMENT,
            title      TEXT    NOT NULL,
            url        TEXT    NOT NULL,
            votes      INTEGER NOT NULL DEFAULT 0,
            created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
        )`)
	return err
}

// insertLink saves a new link and returns it, fully filled in by the database.
func insertLink(title, url string) (Link, error) {
	var l Link
	err := db.QueryRow(`
        INSERT INTO links (title, url)
        VALUES (?, ?)
        RETURNING id, title, url, votes, created_at`,
		title, url,
	).Scan(&l.ID, &l.Title, &l.URL, &l.Votes, &l.CreatedAt)
	return l, err
}

// listLinks returns every link. sort is "top" (most votes) or anything else
// for newest-first. The user's sort choice only *picks between* our two
// hard-coded ORDER BY strings — it never becomes part of the SQL itself.
func listLinks(sort string) ([]Link, error) {
	orderBy := "id DESC" // the default: newest first
	if sort == "top" {
		orderBy = "votes DESC, id DESC"
	}

	rows, err := db.Query(`
        SELECT id, title, url, votes, created_at
        FROM links
        ORDER BY ` + orderBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []Link{}
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.ID, &l.Title, &l.URL, &l.Votes, &l.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

// voteLink adds one upvote to a link and returns its new total.
// It returns sql.ErrNoRows when no link has that id.
func voteLink(id int64) (int64, error) {
	var votes int64
	err := db.QueryRow(`
        UPDATE links SET votes = votes + 1
        WHERE id = ?
        RETURNING votes`,
		id,
	).Scan(&votes)
	return votes, err
}

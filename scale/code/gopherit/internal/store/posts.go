package store

import (
	"database/sql"
	"errors"
)

// Post is one submission in a subreddit. Body and URL are omitempty so a
// link post doesn't show "body": "" and a text post doesn't show "url": "".
type Post struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Body         string `json:"body,omitempty"`
	URL          string `json:"url,omitempty"`
	Score        int64  `json:"score"`
	CommentCount int64  `json:"comment_count"`
	Author       string `json:"author"`
	Subreddit    string `json:"subreddit"`
	CreatedAt    string `json:"created_at"`
}

// PostFilters carries everything a feed query needs: which community
// (0 = all of them), how to sort, and which page to return.
type PostFilters struct {
	SubredditID int64
	Sort        string // "new", "top" or "hot"
	Page        int
	PageSize    int
}

// Meta describes the page of results that came back, so clients can
// render pagination controls.
type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

// CreatePost inserts a post and returns it fully hydrated (with author
// and subreddit names attached).
func (s *Store) CreatePost(subredditID, authorID int64, title, body, url string) (*Post, error) {
	var id int64
	err := s.db.QueryRow(`
        INSERT INTO posts (subreddit_id, author_id, title, body, url)
        VALUES (?, ?, ?, ?, ?)
        RETURNING id`,
		subredditID, authorID, title, body, url,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return s.PostByID(id)
}

// PostByID fetches a single post. Returns ErrNotFound if it doesn't exist.
func (s *Store) PostByID(id int64) (*Post, error) {
	var p Post
	err := s.db.QueryRow(`
        SELECT p.id, p.title, p.body, p.url, p.score, p.created_at,
               u.username, sr.name,
               (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id)
        FROM posts p
        JOIN users      u  ON u.id  = p.author_id
        JOIN subreddits sr ON sr.id = p.subreddit_id
        WHERE p.id = ?`,
		id,
	).Scan(&p.ID, &p.Title, &p.Body, &p.URL, &p.Score, &p.CreatedAt,
		&p.Author, &p.Subreddit, &p.CommentCount)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ListPosts returns one page of the feed plus pagination metadata.
//
// The "hot" sort is the fun part: score divided by (age in hours + 2)
// raised to 1.5 — fresh posts with a few votes beat old posts with many.
// The +2 stops brand-new posts from dividing by (almost) zero.
func (s *Store) ListPosts(f PostFilters) ([]Post, Meta, error) {
	orderBy := map[string]string{
		"new": "p.id DESC",
		"top": "p.score DESC, p.id DESC",
		"hot": "p.hotness DESC, p.id DESC",
	}[f.Sort]
	if orderBy == "" {
		orderBy = "p.id DESC"
	}

	// hotness is now a precomputed column (refreshed by a background goroutine),
	// so the feed no longer computes pow() per row.
	//
	// We dropped the old COUNT(*) OVER() window function. It looked convenient —
	// data + total in one query — but it forced SQLite to scan and sort ALL
	// matching posts on every request (EXPLAIN showed "SCAN p" + "USE TEMP B-TREE
	// FOR ORDER BY"). Without it, the query walks idx_posts_hotness and stops
	// after one page. We fetch the total separately, with a cheap count.
	query := `
        SELECT p.id, p.title, p.body, p.url, p.score, p.created_at,
               u.username, sr.name,
               (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id)
        FROM posts p
        JOIN users      u  ON u.id  = p.author_id
        JOIN subreddits sr ON sr.id = p.subreddit_id
        WHERE (? = 0 OR p.subreddit_id = ?)
        ORDER BY ` + orderBy + `
        LIMIT ? OFFSET ?`

	offset := (f.Page - 1) * f.PageSize
	rows, err := s.db.Query(query, f.SubredditID, f.SubredditID, f.PageSize, offset)
	if err != nil {
		return nil, Meta{}, err
	}
	defer rows.Close()

	posts := []Post{}
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.Title, &p.Body, &p.URL, &p.Score, &p.CreatedAt,
			&p.Author, &p.Subreddit, &p.CommentCount)
		if err != nil {
			return nil, Meta{}, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, Meta{}, err
	}

	// The total is a count-only query (no joins, no sort) — cheap next to
	// materialising every row.
	var total int64
	err = s.db.QueryRow(
		`SELECT COUNT(*) FROM posts WHERE (? = 0 OR subreddit_id = ?)`,
		f.SubredditID, f.SubredditID).Scan(&total)
	if err != nil {
		return nil, Meta{}, err
	}

	meta := Meta{
		Page:       f.Page,
		PageSize:   f.PageSize,
		Total:      total,
		TotalPages: (total + int64(f.PageSize) - 1) / int64(f.PageSize),
	}
	return posts, meta, nil
}

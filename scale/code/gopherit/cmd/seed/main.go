// Command seed fills a gopherit database with realistic fake data so we have
// something big enough to load-test against. It writes SQL directly (not through
// the API) so it finishes in seconds instead of hours:
//   - it hashes ONE password with bcrypt and reuses it for every fake user;
//   - it inserts rows in batches of a few hundred per statement (multi-row
//     VALUES), which is dramatically faster than one INSERT per row.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"gopherit/internal/store"
)

const batchSize = 400 // rows per multi-row INSERT

func main() {
	var (
		dsn      = flag.String("dsn", "gopherit.db", "SQLite database file")
		nUsers   = flag.Int("users", 10000, "how many users")
		nSubs    = flag.Int("subs", 200, "how many subreddits")
		nPosts   = flag.Int("posts", 50000, "how many posts")
		nVotes   = flag.Int("votes", 500000, "how many votes (deduped)")
		seedNum  = flag.Int64("seed", 1, "PRNG seed for repeatable data")
		password = flag.String("password", "password123", "shared password for every seeded user")
	)
	flag.Parse()

	// Ensure the schema exists (Open runs the migration), then use our own
	// connection so we can batch-insert freely.
	st, err := store.Open(*dsn)
	if err != nil {
		log.Fatal(err)
	}
	st.Close()

	db, err := sql.Open("sqlite", *dsn+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rng := rand.New(rand.NewSource(*seedNum))
	start := time.Now()

	hash, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("seeding %d users...", *nUsers)
	users := &batcher{tx: tx, prefix: "INSERT INTO users (username, email, password_hash) VALUES ", cols: 3}
	for i := 1; i <= *nUsers; i++ {
		users.add(fmt.Sprintf("user%d", i), fmt.Sprintf("user%d@example.com", i), hash)
	}
	users.flush()

	log.Printf("seeding %d subreddits...", *nSubs)
	subs := &batcher{tx: tx, prefix: "INSERT INTO subreddits (name, title, creator_id) VALUES ", cols: 3}
	for i := 1; i <= *nSubs; i++ {
		subs.add(fmt.Sprintf("sub%d", i), fmt.Sprintf("Subreddit number %d", i), rng.Intn(*nUsers)+1)
	}
	subs.flush()

	// Posts spread across the last 30 days so "hot" has something to chew on.
	log.Printf("seeding %d posts...", *nPosts)
	now := time.Now().UTC()
	posts := &batcher{tx: tx, prefix: "INSERT INTO posts (subreddit_id, author_id, title, body, created_at) VALUES ", cols: 5}
	for i := 1; i <= *nPosts; i++ {
		created := now.Add(-time.Duration(rng.Intn(30*24)) * time.Hour).Format(time.RFC3339)
		posts.add(rng.Intn(*nSubs)+1, rng.Intn(*nUsers)+1,
			fmt.Sprintf("Post %d about something interesting", i), "Body text for a seeded post.", created)
	}
	posts.flush()

	// Votes: random (user, post) pairs, deduped by the composite primary key.
	log.Printf("seeding ~%d votes...", *nVotes)
	votes := &batcher{tx: tx, prefix: "INSERT OR IGNORE INTO votes (user_id, target_type, target_id, value) VALUES ", cols: 4}
	for i := 0; i < *nVotes; i++ {
		value := 1
		if rng.Intn(5) == 0 { // 1 in 5 is a downvote
			value = -1
		}
		votes.add(rng.Intn(*nUsers)+1, "post", rng.Intn(*nPosts)+1, value)
	}
	votes.flush()

	// Fold the votes into each post's cached score column. We create a TEMPORARY
	// index just for this aggregation and drop it again — the whole point of the
	// book is that gopherit ships WITHOUT a votes index, so the seeded database
	// must not contain one either. (Without the temp index this single UPDATE
	// takes hours: it scans all 500k votes once per post. That slowness is
	// exactly the lesson of Chapter 4.)
	log.Printf("computing scores...")
	if _, err := tx.Exec(`CREATE INDEX seed_tmp_votes ON votes(target_type, target_id)`); err != nil {
		log.Fatal(err)
	}
	if _, err := tx.Exec(`
		UPDATE posts SET score = COALESCE((
			SELECT SUM(value) FROM votes
			WHERE target_type = 'post' AND target_id = posts.id
		), 0)`); err != nil {
		log.Fatal(err)
	}
	if _, err := tx.Exec(`DROP INDEX seed_tmp_votes`); err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	var posts2, votes2 int
	db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&posts2)
	db.QueryRow(`SELECT COUNT(*) FROM votes`).Scan(&votes2)
	log.Printf("done in %s: %d users, %d subs, %d posts, %d votes",
		time.Since(start).Round(time.Millisecond), *nUsers, *nSubs, posts2, votes2)
}

// batcher accumulates rows and flushes them as multi-row INSERT statements.
type batcher struct {
	tx     *sql.Tx
	prefix string
	cols   int
	args   []any
	rows   int
}

func (b *batcher) add(vals ...any) {
	b.args = append(b.args, vals...)
	b.rows++
	if b.rows >= batchSize {
		b.flush()
	}
}

func (b *batcher) flush() {
	if b.rows == 0 {
		return
	}
	one := "(" + strings.TrimRight(strings.Repeat("?,", b.cols), ",") + ")"
	placeholders := strings.TrimRight(strings.Repeat(one+",", b.rows), ",")
	if _, err := b.tx.Exec(b.prefix+placeholders, b.args...); err != nil {
		log.Fatal(err)
	}
	b.args = b.args[:0]
	b.rows = 0
}

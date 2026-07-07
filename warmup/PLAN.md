# Implementation Plan: "linkboard" — the Go REST warm-up book

**Status: PLAN — not yet implemented.** This document is a complete, self-contained
spec for building the warm-up mini-book. Whoever implements it (human or AI agent)
should be able to work from this file alone. Delete or archive this file once the
book ships.

---

## 1. Context & goal

This repo (`devthedeveloper/gocheatsheet`) is a Go learning site served by GitHub
Pages from the `main` branch root: https://devthedeveloper.github.io/gocheatsheet

It already contains:

- `index.html`, `advanced.html`, `programs.html`, `concurrency.html` — the cheatsheet pages
- `ebook/` — **"Go + REST — Build Your Own Reddit"**, a 12-chapter handwritten-notes-style
  ebook that builds `gopherit`, a Reddit-clone REST API (complete tested code in
  `ebook/code/gopherit/`)

**Problem:** readers coming straight from the cheatsheet find the Reddit book's code
too hard — the jump from 15-line hello-world servers to a layered, production-hardened
project is steep.

**Solution:** a 5-chapter **warm-up book** that builds **linkboard** — a tiny anonymous
Hacker-News-style link board (~180 lines total). It teaches the core request loop
(route → decode → query → respond) on a drastically smaller surface, in the same
domain (posts + votes), so the Reddit book afterwards reads like review, not a wall.

The learning path becomes: **cheatsheet → 🐣 warm-up (linkboard) → 📕 Reddit book (gopherit)**.

## 2. The app: linkboard

One shared board. Anyone can post a link and upvote. Deliberately **excluded**
(these are the Reddit book's job): users, auth, tokens, middleware, comments,
communities, pagination, transactions, testing, layered packages.

### Final endpoints

| Method & path            | Body / query                          | Response |
|--------------------------|---------------------------------------|----------|
| `GET /links`             | `?sort=new` (default) or `?sort=top`  | `200` JSON array of links |
| `POST /links`            | `{"title": "...", "url": "https://..."}` | `201` the created link |
| `POST /links/{id}/vote`  | (empty body)                          | `200` `{"id": N, "votes": M}`; `404` if unknown id |

Link JSON shape: `{"id": 1, "title": "...", "url": "...", "votes": 3, "created_at": "..."}`.

Error shape (from ch4 on): `{"error": "message"}` with correct status codes
(400 bad JSON, 404 unknown id, 422 failed validation).

### Final file layout (~180 lines total, single package)

```
linkboard/
├── go.mod              # module linkboard; only dep: modernc.org/sqlite
├── main.go             # main(), mux, server
├── store.go            # openDB + the 4 SQL functions
├── handlers.go         # the 3 handlers + tiny helpers
└── index.html          # ch4: 25-line vanilla-JS frontend (served by the API)
```

Keep it ONE package (`main`), NO `internal/`, NO application struct, NO middleware,
NO helper layers beyond one `writeJSON` and one `writeError` (≤8 lines each).
Plain functions, package-level `db *sql.DB` variable is acceptable here — the
Reddit book explains later why bigger apps avoid it (do NOT pre-optimize this book).

### SQLite schema (chapter 2)

```sql
CREATE TABLE IF NOT EXISTS links (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    title      TEXT    NOT NULL,
    url        TEXT    NOT NULL,
    votes      INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
```

Driver: `modernc.org/sqlite` (pure Go, no cgo), DSN `linkboard.db`. Keep pragmas to
just `?_pragma=busy_timeout(5000)` — foreign keys/WAL aren't needed for one table
and shouldn't be explained here.

## 3. Chapter-by-chapter spec

Each chapter ends with a **checkpoint**: a complete, compiling folder under
`warmup/code/checkpoints/ch0N/` reflecting the project exactly as of that chapter's
end. Chapter 1's checkpoint is one file; ch05's equals the final app.
`warmup/code/checkpoints/go.work` or per-checkpoint `go.mod` files — use ONE
`go.mod` per checkpoint folder (module name `linkboard`), so each is runnable with
`go run .` inside its folder.

### Ch 1 — One file, whole API (~60 lines)

- `main.go` only. Links in a package-level `[]Link` slice, next-id counter.
- `GET /links` and `POST /links` via `http.NewServeMux` with Go 1.22+ patterns.
- A 3-line `sync.Mutex` around slice access, with a sticky note: *why* (one goroutine
  per request) + link to the cheatsheet's concurrency page. Keep it to 3 sentences.
- Reader payoff: working API + curl transcript within ~15 minutes.
- Teaches: handler signature, mux patterns, `json.NewEncoder/Decoder`, struct tags,
  201 status.

### Ch 2 — Survive a restart (SQLite)

- Motivation hook: "restart the server — your links are gone."
- Add `store.go`: `openDB()`, `insertLink`, `listLinks`. Delete the slice + mutex
  (the DB handles concurrent access — one reassuring sentence, no deep dive).
- `go get modernc.org/sqlite`, blank-import explanation (3 sentences max, point to
  Reddit book ch5 for depth).
- Teaches: `sql.Open`, `Exec`, `Query`/`Scan` row loop (defer Close, rows.Err),
  `QueryRow(...RETURNING id)`.

### Ch 3 — Votes & sorting

- `POST /links/{id}/vote`: `r.PathValue("id")` + strconv; SQL
  `UPDATE links SET votes = votes + 1 WHERE id = ? RETURNING votes` → 404 on
  `sql.ErrNoRows`.
- `?sort=top|new`: validated against a 2-entry map choosing between two hard-coded
  `ORDER BY` strings (one warn box: user input never becomes SQL — it only *selects*
  between our strings).
- Anonymous votes on purpose — a sticky notes that "one vote per user" needs
  accounts, which is exactly what the Reddit book adds.

### Ch 4 — Be a good API (+ the 30-second frontend)

- Add `writeJSON(w, status, v)` and `writeError(w, status, msg)` — each ≤8 lines.
- Validation: title non-empty & ≤200 chars; url starts with `http://`/`https://`
  → 422 with `{"error": "..."}`. Bad JSON → 400. (Simple `if` checks, NO validator
  type — that's the Reddit book's upgrade.)
- One line of CORS (`Access-Control-Allow-Origin: *`) + 2 sentences on why.
- `index.html`: ~25 lines of vanilla JS (fetch list, render `<li>`s, form POST,
  vote buttons). Served with `mux.Handle("GET /", http.FileServer(...))` or
  `http.ServeFile` — pick the simplest that works. The "I built a website" moment.

### Ch 5 — Ship it + the bridge

- `go build`, run the binary, 3-line cross-compile mention.
- **The bridge** (this chapter's real payload): a table mapping every linkboard
  concept to its grown-up gopherit counterpart, e.g. anonymous votes → users +
  one-vote-per-user upserts (Reddit ch6/11); one table → six tables with foreign
  keys (ch5); inline validation → problems validator (ch4); `writeError` → error
  taxonomy (ch4); no auth → bearer tokens (ch7); etc. End with a link straight into
  `../ebook/index.html`.

## 4. Book style & mechanics (match the existing ebook exactly)

Study 2–3 chapters of `ebook/` first (`ebook/ch02.html` and `ebook/ch09.html` are
good references) and reuse the same components and voice:

- **Shared assets, do not duplicate:** each warmup page links
  `../ebook/fonts/fonts.css`, `../ebook/notes.css`, `../ebook/notes.js`
  (self-hosted fonts already live in `ebook/fonts/`).
- Components: `.sheet`, `booknav` top/bottom, `.chapter-head` + `.ch-sticker`,
  "In this chapter you'll learn" `.sticky`, `.swipe` h2 highlights, `.code-card`
  with `.filename` labels, dark `.term-card` for terminal transcripts, boxes
  (`.box.tip/.warn/.gotcha/.deep`), `.flow` diagrams, `.scribble`, `.recap`
  "Pin to your brain" checklists.
- Tone: friendly, colorful, analogies, second person; every chapter ends with a
  recap + a one-line teaser for the next.
- Pages to create: `warmup/index.html` (small cover + 5 TOC cards + "who this is
  for") and `warmup/ch01.html` … `warmup/ch05.html`.

### Non-negotiable correctness rules (lessons already paid for)

1. **Every code card shows a complete file** — including `package main` and imports.
   Never a floating function labelled as a new file. (A reader already hit
   `expected 'package', found 'func'` from the Reddit book doing this; it was fixed
   there — don't reintroduce it.)
2. **Write and verify the code FIRST, then write chapters around it.** Every
   checkpoint must pass `go vet ./...` and `go build ./...`, and be exercised with
   real `curl` runs. Paste the *actual* captured outputs into the term-cards —
   never invent output.
3. If code appears in both a checkpoint and a chapter, the chapter's copy must be
   byte-identical to the file on disk (inject programmatically or copy carefully,
   HTML-escaping `&`, `<`, `>`; the ebook used `data-src` attributes on
   `<code>` elements + a small injection script — reusing that pattern is
   recommended).
4. If a headless Chromium is available (`/opt/pw-browsers/chromium` in Claude Code
   web sandboxes), screenshot each page at ~1100px wide and eyeball it before
   pushing.

## 5. Site integration

1. **Cheatsheet chips** (all 4 pages: `index.html`, `advanced.html`,
   `programs.html`, `concurrency.html`): add `<a class="chip swap"
   href="warmup/">🐣 Warm-up: first API</a>` immediately BEFORE the existing
   `📕 REST ebook` chip, so the chips read as a path.
2. **Reddit book cover** (`ebook/index.html`): add one sticky near "Before you
   start": *"First API ever? Do the 🐣 warm-up book first — you'll build a tiny
   link board in an evening, and this book will feel far easier."* linking to
   `../warmup/`.
3. **Reddit book ch1**: optional single-line pointer to the warm-up for readers
   who feel lost. Keep it to one sentence.
4. `warmup/README.md`: short — what it is, how to read, where the code lives
   (mirror `ebook/README.md`'s shape).

## 6. Delivery & verification checklist

- [ ] `warmup/code/checkpoints/ch01..ch05/` all pass `go vet` + `go build`; ch05 is the complete app
- [ ] Each checkpoint smoke-tested with curl; transcripts captured for the chapters
- [ ] ch04's `index.html` frontend actually works against the running API (verify in headless browser if available)
- [ ] 6 HTML pages written (`index` + 5 chapters), styled identically to `ebook/`
- [ ] All internal links work: warmup ↔ ebook ↔ cheatsheet, prev/next navs, TOC cards
- [ ] Chips added on all 4 cheatsheet pages; sticky added on ebook cover
- [ ] Everything committed to `main` and pushed (this repo deploys Pages from `main` root — pushing = publishing)
- [ ] Verify the Pages "pages build and deployment" workflow run succeeds for the commit (the sandbox proxy usually can't fetch `github.io` directly — check the workflow status via the GitHub API instead)
- [ ] Delete or archive this PLAN.md in the final commit

## 7. Out of scope (do not add)

Auth of any kind, middleware chapters, pagination, comment threads, tests chapter,
Docker, deployment guides beyond `go build`, downvotes, edit/delete endpoints,
config flags (hardcode port 4000), logging beyond `log.Println` at startup.
The warm-up's entire value is what it leaves out. If a feature feels tempting,
it belongs in the Reddit book instead.

## 8. Sizing

~180 lines of final Go, 5 checkpoints, 6 HTML pages of ~150–300 lines each.
Expect the implementation order: code + checkpoints (verify) → chapters → cover →
integration links → screenshots → push → confirm Pages build.

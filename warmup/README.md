# Build Your First Go API — the linkboard warm-up 🐣

A gentle, 5-chapter mini-book that builds **linkboard**: a tiny anonymous
Hacker-News-style link board (~180 lines of Go). It's the stepping stone to the
[Reddit book](../ebook/) next door — same handwritten-notes style, a fraction of
the difficulty.

## Read it

Open [`index.html`](index.html) in a browser (that's the cover + table of
contents), or start live at
https://devthedeveloper.github.io/gocheatsheet/warmup/

## Chapters

1. **One File, Whole API** — links in memory; handlers, routing, JSON
2. **Survive a Restart** — SQLite in one file; `Exec` / `QueryRow` / row-loop
3. **Votes & Sorting** — `UPDATE ... RETURNING`, path params, `?sort=top|new`
4. **Be a Good API + a Web Page** — JSON errors, validation, CORS, a 30-line frontend
5. **Ship It + the Bridge** — one binary, and a map to every idea in the Reddit book

## The code

[`code/checkpoints/ch01..ch05/`](code/checkpoints/) — a complete, compiling copy
of the project as of the end of each chapter. Each folder runs with `go run .`
inside it (Go 1.24+; only dependency is `modernc.org/sqlite`). Every code block
in the book is injected from these files, so book and code never drift apart.

Stuck at any point? Copy the matching checkpoint folder and keep going.

## Then what?

Do the [**Go + REST — Build Your Own Reddit**](../ebook/) book. Everything here
is the seed of everything there — chapter 5 draws the full map.

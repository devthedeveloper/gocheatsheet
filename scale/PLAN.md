# Implementation Plan: "Go + Scale — Survive the Front Page" (book 3)

**Status: PLAN — not yet implemented.** Self-contained spec for the third book in
the series. Whoever implements it (human or AI agent) should be able to work from
this file alone. Archive it once the book ships.

---

## 1. Context & goal

This repo (`devthedeveloper/gocheatsheet`) is a Go learning site on GitHub Pages
(`main` branch root → https://devthedeveloper.github.io/gocheatsheet). It hosts:

- the cheatsheet pages (`index.html`, `advanced.html`, `programs.html`, `concurrency.html`)
- `warmup/` — **Build Your First Go API** (linkboard, 5 chapters)
- `ebook/` — **Go + REST — Build Your Own Reddit** (gopherit, 12 chapters,
  complete tested code in `ebook/code/gopherit/`)

The learning path is: **cheatsheet → 🐣 warm-up → 📕 Reddit book → (this book)**.

**The question this book answers:** *"gopherit works. Now imagine millions of
users — will the server cope?"* Readers finish the Reddit book with a working
backend and no idea what happens under load, when it breaks, or what to fix
first. Most scaling material is architecture astrology — diagrams of load
balancers and Redis with no numbers. This book's thesis is the opposite:

> **Scaling is measurement.** Load-test it, watch it break, fix the actual
> bottleneck, measure again. Repeat until the front page can't kill you.

Working title: **"Go + Scale — Survive the Front Page"**. Directory: `scale/`.
The "hug of death" (your post hits #1 and the crowd arrives) is the running
story.

## 2. The starting point & the four cracks

The book begins from the finished gopherit (`ebook/code/gopherit/`, copied into
this book's code dir — do not modify the ebook's copy). Under load it cracks in
four places, each deliberately planted by the earlier book's "keep it simple"
choices — this book pays off those IOUs:

1. **`db.SetMaxOpenConns(1)`** — the training-wheels setting. One connection for
   ALL reads and writes serializes everything.
2. **Token lookup per authenticated request** — `SELECT` on every request with a
   Bearer header.
3. **bcrypt on `/tokens`** — ~250ms per login *by design* → a free DoS button.
4. **The hot feed** — `pow()` per row per request, no covering index, no cache,
   on the most-hit endpoint of the site.

Also inherited on purpose: store methods take no `context.Context` (the ebook
explicitly deferred this — chapter 6 here pays it off).

## 3. Teaching principles (non-negotiable)

1. **Every number in the book is real.** All benchmark tables/transcripts come
   from actual runs by the implementer. Present as "on my machine" with the
   exact reproduction commands. Never invent or extrapolate numbers. Relative
   improvements (2.4×) are the lesson, not absolute rps.
2. **Measure → hypothesize → fix → re-measure** is the loop every chapter walks.
   No fix lands without a before/after.
3. **Build the tools first, adopt industry tools second.** The load tester is
   ~120 lines of Go the reader writes (goroutines/channels/atomics — ties to the
   concurrency cheatsheet). A box mentions k6/vegeta/wrk as the grown-up tools.
4. **Same style, same rules as the other books:** handwritten-notes theme via
   `../ebook/notes.css` + fonts + notes.js (shared, not duplicated); complete
   files in code cards (never a floating function — include `package` lines);
   per-chapter checkpoint folders that compile; code injected byte-identically
   from checkpoints (reuse `warmup/inject.py` pattern with a `scale/` root).
5. **Honesty about scale:** an early "napkin math" section shows millions of
   *users* ≠ millions of *concurrent requests* (1M DAU ≈ ~600 rps avg, few-k
   peak, >90% reads) — and a final chapter is honest that most sites never need
   more than this book teaches.

## 4. The code

```
scale/code/
├── gopherit/            # the evolving app (starts as a copy of ebook's final code)
├── hammer/              # the ~120-line Go load tester built in ch2
├── checkpoints/ch01..ch11/   # full compiling snapshots per chapter (warmup pattern)
└── bench/               # reproducible benchmark scripts (bash) used for every table
```

Dependency policy (keep the series' minimalism, relax only where the lesson
demands): `golang.org/x/sync/singleflight` (ch5), `golang.org/x/time/rate`
(ch7), a Postgres driver `github.com/jackc/pgx/v5/stdlib` (ch8),
`github.com/prometheus/client_golang` (ch10). Redis appears only as a short
"here's the 20 lines" section with the in-process path remaining the fully
working default (see ch9) — implementer may verify Redis code with
`miniredis` in a test if no server is available.

**Postgres verification strategy:** the store grows a `-db=sqlite|postgres` flag
and BOTH paths stay working (great design lesson: the store interface is the
seam). If the authoring sandbox has no Postgres, chapter 8's code must still be
compile-verified and its transcripts captured against Postgres wherever the
implementer CAN run one (local `docker run postgres` one-liner is in the
chapter); the SQLite path keeps every checkpoint runnable everywhere.

## 5. Chapter-by-chapter spec (11 chapters)

### Ch 1 — The Hug of Death (no code)
The story: your post hits #1. Napkin math from "millions of users" to rps
(DAU × req/day ÷ 86 400 × peak factor; read/write ratio ~95/5 for a Reddit).
Throughput vs latency; why averages lie; meet p50/p95/p99 (queue-at-the-café
analogy). The measurement loop diagram. Honest sticky: "a single Go box is
faster than you think — we'll prove it, then break it."

### Ch 2 — Build a Load Cannon
Write `hammer`: N goroutine workers, shared atomic counters, latency histogram,
`-rps`, `-duration`, `-url`, mixed read/write scenarios (weighted endpoints:
feed-heavy, some logins, some votes). This is applied concurrency-cheatsheet
material. Then: **baseline gopherit** (seeded DB: ~10k users, 100 subreddits,
50k posts, 500k votes — seeding script included). Capture the first crash/cliff:
where writes stall on the single connection, where p99 explodes. A box on
k6/vegeta/wrk. All numbers real.

### Ch 3 — Find the Bottleneck (profiling)
`net/http/pprof` (one import on our mux — explain the guard: never expose it
publicly), CPU + heap profiles under load, reading a flame graph,
`EXPLAIN QUERY PLAN` on the feed query (see the full scan / temp b-tree),
`go build -race` under load as the correctness check. Teach "the bottleneck is
never where you guessed" — show what the profile ACTUALLY says before fixing
anything.

### Ch 4 — Database Surgery
Fixes driven by ch3's evidence, re-benchmarked one at a time:
(a) proper indexes for the feed & token lookup — watch EXPLAIN flip;
(b) undo `SetMaxOpenConns(1)`: WAL allows concurrent readers + one writer —
    size the pool, keep writes serialized (SQLite reality box);
(c) stop computing hotness per row per request: a `hotness` column recomputed
    by a background goroutine every 30s (first taste of "do less work per
    request" + a ticker goroutine). Honest box: the score is now ≤30s stale —
    nobody can tell, and THAT tradeoff is the essence of scaling.

### Ch 5 — Cache Rules Everything
The front page is the same bytes for everyone — stop rebuilding it. Build a
tiny in-process TTL cache (map + RWMutex, ~40 lines) for feed pages; the
thundering-herd problem and `singleflight` (one flight rebuilds, everyone
shares); invalidation choices (TTL vs explicit on new-post/vote — pick TTL=1s,
explain why). Re-benchmark: this is the chapter with the jaw-drop read numbers.
Box: HTTP-level caching (Cache-Control/ETag) exists too; one paragraph.

### Ch 6 — Contexts & Timeouts (paying the ebook's IOU)
Thread `r.Context()` into every store method (the mechanical refactor is shown
once, then "repeat for the rest" with the checkpoint as reference). Query
timeouts with `context.WithTimeout`; client-disconnect cancellation demoworthy
with a slow query + curl Ctrl+C; server-wide time budgets. Sticky: "a request
that can't be cancelled is a leak under load."

### Ch 7 — The Login Problem (rate limiting)
bcrypt is slow on purpose → `/tokens` is a DoS magnet (hammer demo: a few
hundred rps of logins pins all cores). Fixes: per-IP token-bucket middleware
(`x/time/rate`, limiter map with janitor goroutine — the ebook's quest-log item
done properly), plus an in-process auth cache (token-hash → user, short TTL) so
hot requests skip the per-request SELECT. Re-benchmark both.

### Ch 8 — Graduate to Postgres
When SQLite's single-writer ceiling is *actually* the bottleneck (show it), the
store-layer promise pays off: add `-db=postgres`, swap driver to pgx/stdlib,
`?` → `$1` placeholders, connection-pool sizing (SetMaxOpenConns/Idle/
ConnMaxLifetime rules of thumb), the docker one-liner to run PG locally, and
the migration honesty box (moving data, and what we'd use goose for). Handlers
change ZERO lines — make that the loudest sentence in the book. Re-benchmark
writes: concurrent writers unlocked.

### Ch 9 — Clones (scale out)
Stateless payoff: run 3 gopherit instances (ports 4001-3) behind a local
round-robin proxy (Caddy config, ~5 lines; box explains any LB is the same
idea). What breaks when you clone: the in-process cache and rate limiter are
now per-instance — introduce Redis as the *shared* versions of both (short,
~20 lines each, clearly optional; in-process remains the default path).
Vertical vs horizontal box; session-affinity myth debunked (we never had
in-memory sessions — by design, since book 2 ch7).

### Ch 10 — Watch It Breathe (observability)
You can't fix what you can't see at 3 a.m.: Prometheus `/metrics` endpoint,
request-duration histogram middleware (one more onion layer — the pattern from
book 2 pays again), the four golden signals, what to alert on (p99 + error
rate, not CPU), and a plain-text box on dashboards (Grafana screenshot optional
— only if actually produced; never fake one). `expvar` mention as the
stdlib-only alternative.

### Ch 11 — The Rematch + the Map
Re-run chapter 2's full scenario against the final system: the whole-book
before/after table, chapter by chapter, with the honest ranking of what
mattered (spoiler the data will likely show: cache ≥ indexes > pool sizing >
everything else). Then "the map of what we didn't need": CDN, sharding, queues,
microservices, Kubernetes, multi-region — one paragraph each on WHEN you'd
know you need it (with the trigger metric). Further reading: *Designing
Data-Intensive Applications*, the Go pprof docs, use-the-index-luke. Close the
trilogy: cheatsheet → linkboard → gopherit → gopherit-at-scale.

## 6. Book style & site integration

- Pages: `scale/index.html` (cover: gopher with a hard hat / flame background
  motif in the existing SVG-doodle style, TOC cards, "who this is for" =
  finished book 2 or equivalent experience) + `ch01.html`…`ch11.html`.
- Shared assets via `../ebook/` links exactly like `warmup/` does.
- New small CSS additions if needed (e.g. a benchmark-table style) go in a
  `scale/scale.css` extension file, not by editing `notes.css`.
- Benchmark tables: use `table.notes` with `font-variant-numeric: tabular-nums`;
  always show the command that produced them in a term-card directly above.
- Chips on all four cheatsheet pages: add `<a class="chip swap" href="scale/">🚀
  Scale it</a>` immediately AFTER the `📕 REST ebook` chip (path order:
  🐣 → 📕 → 🚀). Add a "finished the book? scale it" sticky/link on
  `ebook/ch12.html` (in "Your quest log" or the closing section) and a one-line
  pointer on `ebook/index.html`.
- `scale/README.md` mirroring the other books' READMEs.
- Update `warmup/ch05.html`'s bridge ONLY if trivially safe (optional one-liner
  mentioning book 3 exists); do not restructure it.

## 7. Delivery & verification checklist

- [ ] All checkpoints compile (`go vet` + `go build`) and the seeded-DB script runs
- [ ] `hammer` works; every benchmark table in the book reproduced by `scale/code/bench/*.sh` scripts with outputs captured verbatim
- [ ] pprof/EXPLAIN transcripts are real captures
- [ ] Both `-db=sqlite` and `-db=postgres` paths compile; sqlite path fully tested in-sandbox; postgres path tested wherever a PG is available (document which)
- [ ] gopherit's existing httptest suite still passes at every checkpoint (regressions = the book broke the app)
- [ ] 12 HTML pages styled identically to the existing books; all internal links resolve in both directions (warmup ↔ ebook ↔ scale ↔ cheatsheet)
- [ ] Chips added on 4 cheatsheet pages; pointers added on ebook ch12 + cover
- [ ] Headless screenshots of each page eyeballed before pushing
- [ ] Push to `main`, then confirm the "pages build and deployment" workflow succeeds via the GitHub API (sandbox proxy can't fetch github.io directly)
- [ ] Archive this PLAN.md in the final commit

## 8. Out of scope (do not add)

Kubernetes, microservices, service meshes, gRPC, sharding, multi-region,
Kafka/NATS (queues get one WHEN-you-need-it paragraph in ch11), CDNs beyond a
box, autoscaling, TLS termination details, cloud-provider specifics, frontend
performance. If a topic can't be *measured on the reader's laptop*, it belongs
in ch11's "map", not in a chapter.

## 9. Sizing

~11 checkpoints of a ~20-file app + hammer (~120 lines) + bench scripts;
12 HTML pages (~200–350 lines each). The heavy part is doing the benchmarks
honestly — budget more time for running/capturing than for prose. Expected
order: copy code → seed script → hammer → baseline capture → per-chapter
fix+capture loop → chapters → cover → integration → screenshots → push →
verify Pages.

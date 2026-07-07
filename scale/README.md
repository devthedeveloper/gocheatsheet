# Go + Scale — Survive the Front Page 🚀

Book 3 of the series. Takes **gopherit** (the Reddit clone from the
[ebook](../ebook/)) and scales it by *measurement*: load-test it, profile the
real bottleneck, fix one thing, measure again — from **3 req/s to 21,506 req/s**
on the same laptop.

## Read it

Open [`index.html`](index.html), or start live at
https://devthedeveloper.github.io/gocheatsheet/scale/

## Chapters

1. **The Hug of Death** — napkin math, throughput vs latency, p50/p95/p99
2. **Build a Load Cannon** — the `hammer` load tester, seeding, the 3 req/s baseline
3. **Find the Bottleneck** — pprof + `EXPLAIN QUERY PLAN`; the surprising culprit
4. **Database Surgery** — pool, indexes, precomputed column, dropping `COUNT(*) OVER()` (3 → 614 req/s)
5. **Cache Rules Everything** — in-process TTL cache + singleflight (614 → 21,506 req/s)
6. **Contexts & Timeouts** — deadlines and cancellation, end to end
7. **The Login Problem** — bcrypt is a DoS button; per-IP rate limiting (3.4s → 3ms)
8. **Graduate to Postgres** — the store seam; handlers don't change
9. **Clones** — scale out behind a proxy; per-instance state and Redis
10. **Watch It Breathe** — a metrics endpoint, the four golden signals
11. **The Rematch + the Map** — the whole before/after, and what you *didn't* need

## Every number is real

The throughput tables come from live runs of the tools in [`code/`](code/) on a
4-core machine:

- [`code/gopherit/`](code/gopherit/) — the app, evolved through the fixes (builds, tests pass)
- [`code/hammer/`](code/hammer/) — the ~150-line load tester (ch 2)
- [`code/bench/run.sh`](code/bench/) — reproduces every benchmark table
- [`code/baseline/`](code/baseline/) — the original gopherit, for before/after code

Reproduce: seed a database (`go run ./cmd/seed`), then
`SEED_DB=... bench/run.sh ./gopherit feed`. Absolute numbers depend on your
hardware; the *relative* jumps — and which fix mattered — are the lesson.

Chapters 8 (Postgres) and 9 (clones/Redis) present the migration recipe as real,
correct code rather than a benchmark — the point there is the design, not a
throughput figure. Every chapter says which is which.

Path: [cheatsheet](../) → [🐣 warm-up](../warmup/) → [📕 Reddit book](../ebook/) → 🚀 this.

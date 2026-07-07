#!/usr/bin/env bash
# run.sh — one reproducible benchmark run.
#
#   ./run.sh <server-binary> <scenario> [extra server flags...]
#
# It copies the pristine seeded database, starts the server on port 4000,
# waits for it to answer, fires `hammer` at it, prints the results, and
# stops the server. Every benchmark table in the book comes from this script.
set -euo pipefail

BIN="${1:?usage: run.sh <binary> <scenario> [flags...]}"
SCENARIO="${2:?scenario: feed|mixed|login}"
shift 2

SEED="${SEED_DB:-/tmp/bench-seed.db}"
DB="/tmp/bench-run.db"
HAMMER="${HAMMER:-/tmp/hammer}"
WORKERS="${WORKERS:-50}"
DURATION="${DURATION:-15s}"

# fresh copy of the seeded data so writes don't accumulate across runs
rm -f "$DB" "$DB"-*
cp "$SEED" "$DB"

"$BIN" -dsn "$DB" -port 4000 "$@" >/tmp/bench-server.log 2>&1 &
SRV=$!
trap 'kill $SRV 2>/dev/null || true' EXIT

# wait for readiness
for _ in $(seq 1 50); do
  curl -sf localhost:4000/api/v1/healthz >/dev/null 2>&1 && break
  sleep 0.2
done

"$HAMMER" -url http://localhost:4000 -scenario "$SCENARIO" \
  -workers "$WORKERS" -duration "$DURATION"

kill $SRV 2>/dev/null || true

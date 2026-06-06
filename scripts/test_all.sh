#!/usr/bin/env bash
# Full local validation superset — everything an agent or developer should run
# before merging. Composes existing make targets; does not duplicate their logic.
#
# Phase 1 — make test (no server):
#   validate-shared, test-go, test-py, client-unit
#
# Phase 2 — make ci (throws away Postgres + server for integration):
#   validate-assets, Go tests, Python tests, protocol bot, replay verification,
#   headless bot-client (all client scenarios), client-smoke slice
#
# Phase 3 — make bot-visual (own server; headless replay playlist):
#   record bot scenarios, verify replay manifest, run visual replay playlist
#
# Override GODOT, BASE_URL, etc. on the command line as for individual targets.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "== test-all 1/3: unit tests (make test) =="
make test

echo "== test-all 2/3: CI integration (make ci) =="
make ci

echo "== test-all 3/3: headless bot-visual (make bot-visual) =="
GODOT_FLAGS="${GODOT_FLAGS:---headless}" \
ARPG_VISUAL_REPLAY_EXIT_ON_COMPLETE="${ARPG_VISUAL_REPLAY_EXIT_ON_COMPLETE:-1}" \
make bot-visual

echo "test-all OK"

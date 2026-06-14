# v152 Spec: Bot Context Movement Decouple

Status: Complete
Date: 2026-06-14
Codename: `bot-context-movement-decouple`

## Purpose

Begin paying down the 43 `helpers=globals()` coupling sites that v151 baselined as debt, by
replacing the laundering pattern with a real decoupling primitive and migrating the first module.
v145-v149 moved helper bodies out of `run.py` but kept them coupled: each extracted function
received `run.py`'s entire namespace via `helpers=globals()`, so the modules were not importable or
testable without `run.py`. v151 added a ratchet to stop *new* such sites but froze the existing 43.

This slice introduces:

1. `tools/bot/runtime_queries.py` — pure read-only `RuntimeState` queries (`find_player`,
   `dict_distance`) that any module imports directly.
2. `tools/bot/bot_context.py` — a typed `BotContext` carrying only the *stateful* runtime services a
   module needs (currently `pump_one`), injected explicitly instead of via `globals()`.

and migrates `tools/bot/movement_runtime.py` off `helpers=globals()` to prove the pattern: pure deps
imported directly, the one runtime service taken as `ctx: BotContext`, and the four intra-module
calls (which previously bounced out to `run.py` and back) made direct. The module no longer imports
or depends on `tools.bot.run` and is unit-testable in isolation.

## Non-goals

- No scenario behavior change. This is a pure dependency-rewiring refactor; `make bot` and `make ci`
  remain the behavioral regression proof.
- No migration of the other four runtime modules (`runtime_assertions`, `wait_runtime`,
  `state_ingest`, `coop_runtime`). Those follow under touch-to-shrink as later slices touch them.
- No change to `run.py`'s public wrapper signatures — all scenario call sites stay identical.
- No new `BotContext` fields beyond what movement needs (`pump_one`); fields are added when a
  migrated module requires them.

## Acceptance Criteria

- `tools/bot/movement_runtime.py` imports only leaf modules (`bot_types`, `protocol`,
  `runtime_queries`, `bot_context`) — never `tools.bot.run`, directly or transitively at import time.
- `find_player` and `dict_distance` live in `runtime_queries.py`; `run.py` imports them (re-exporting
  the names so unmigrated modules' `globals()` lookups still resolve).
- Movement functions needing the WebSocket pump take `ctx: BotContext`; pure functions
  (`range_candidate_positions`, `derived_walk_max_ticks`) take no context.
- `run.py`'s movement wrappers build `BotContext(pump_one=pump_one)` via `_runtime_context()` instead
  of passing `helpers=globals()`. The `run.py` coupling count drops from 43 to 37, and
  `.maintainability/extraction-coupling-baseline.tsv` is lowered to 37 in the same slice.
- `tools/bot/test_movement_runtime.py` proves extraction independence: a subprocess import asserts
  `tools.bot.run` is not in `sys.modules` after importing `movement_runtime`, and runtime functions
  are tested with a stub `BotContext` and a directly-built `RuntimeState`.
- `make maintainability` (both ratchets) and `make ci` pass.

## Likely Files

- `tools/bot/runtime_queries.py` (new)
- `tools/bot/bot_context.py` (new)
- `tools/bot/movement_runtime.py`
- `tools/bot/run.py`
- `tools/bot/test_movement_runtime.py` (new)
- `.maintainability/extraction-coupling-baseline.tsv`
- `docs/CODEMAP.md`
- `PROGRESS.md`
- `docs/as-built/v152_bot-context-movement-decouple.md`

## Test And Bot Proof

- `.venv/bin/python -m py_compile tools/bot/run.py tools/bot/movement_runtime.py tools/bot/runtime_queries.py tools/bot/bot_context.py`
- `.venv/bin/python -c "import tools.bot.run"` (no circular import)
- `.venv/bin/python -m pytest tools/bot/test_movement_runtime.py -q`
- `make maintainability`
- `make bot`
- `make ci`

## Open Questions And Risks

- The other four runtime modules still use `helpers=globals()` (37 sites). This slice intentionally
  proves the pattern on one module; the rest are touch-to-shrink follow-ups. The coupling ratchet's
  "target: down" now has a working migration recipe to follow.
- `BotContext` deliberately starts minimal (one field). Resist widening it into a god-context; add a
  field only when a migrated module genuinely needs that runtime service.

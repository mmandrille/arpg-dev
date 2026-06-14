# v146 As-Built: Bot Movement Runtime Split

Date: 2026-06-14
Spec: [`docs/specs/v146_spec-bot-movement-runtime-split.md`](../specs/v146_spec-bot-movement-runtime-split.md)
Plan: [`docs/plans/v146_2026-06-14-bot-movement-runtime-split.md`](../plans/v146_2026-06-14-bot-movement-runtime-split.md)

## What Shipped

- Moved Python protocol bot movement/pathing helpers into `tools/bot/movement_runtime.py`.
- Kept `tools.bot.run` compatibility wrappers for walking, move-to-position, in-range movement,
  movement candidate calculation, derived walk budget, and movement accept/reject waiting.
- Passed existing helper bindings from `run.py` into the extracted module, avoiding a reverse import
  of `tools.bot.run` while `python -m tools.bot.run` executes as `__main__`.
- Lowered the `tools/bot/run.py` maintainability baseline from 4768 to 4612 lines and added the new
  movement runtime module to the CODEMAP Bot / scenarios row.

## Proof

- `python -m py_compile tools/bot/run.py tools/bot/movement_runtime.py`
- `make test-py`
- `make bot`
- `.venv/bin/python tools/validate_codemap.py`
- `make maintainability`
- `make ci`

## Deferred

- Splitting action dispatch, wait/pump helpers, state ingestion, co-op orchestration, and replay
  helpers out of `tools/bot/run.py` remains future paydown.

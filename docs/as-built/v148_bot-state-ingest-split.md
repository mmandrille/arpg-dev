# v148 As-Built: Bot State Ingest Split

Date: 2026-06-14
Spec: [`docs/specs/v148_spec-bot-state-ingest-split.md`](../specs/v148_spec-bot-state-ingest-split.md)
Plan: [`docs/plans/v148_2026-06-14-bot-state-ingest-split.md`](../plans/v148_2026-06-14-bot-state-ingest-split.md)

## What Shipped

- Moved Python protocol bot snapshot/delta ingestion into `tools/bot/state_ingest.py`.
- Moved related runtime state mutation helpers for teleporter parsing, hotbar/inventory/stash
  mutation, cooldown decay, active-level clearing, initial-position tracking, and runtime distance
  tracking into the same focused module.
- Kept `tools.bot.run` compatibility wrappers for the moved helper names so existing bot flow and
  previously extracted modules continue to work.
- Lowered the `tools/bot/run.py` maintainability baseline from 4546 to 4288 lines and added the new
  state ingestion module to the CODEMAP Bot / scenarios row.

## Proof

- `python -m py_compile tools/bot/run.py tools/bot/state_ingest.py`
- `make test-py`
- `make bot`
- `.venv/bin/python tools/validate_codemap.py`
- `make maintainability`
- `make ci`

## Deferred

- Splitting action dispatch, co-op orchestration, and replay helpers out of `tools/bot/run.py`
  remains future paydown.

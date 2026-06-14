# v147 As-Built: Bot Wait Runtime Split

Date: 2026-06-14
Spec: [`docs/specs/v147_spec-bot-wait-runtime-split.md`](../specs/v147_spec-bot-wait-runtime-split.md)
Plan: [`docs/plans/v147_2026-06-14-bot-wait-runtime-split.md`](../plans/v147_2026-06-14-bot-wait-runtime-split.md)

## What Shipped

- Moved Python protocol bot wait/pump helpers into `tools/bot/wait_runtime.py`.
- Kept `tools.bot.run` compatibility wrappers for accepts, rejects, events, progression/cooldown
  waits, level/teleporter waits, player-position waits, and WebSocket message pumping.
- Passed existing helper bindings from `run.py` into the extracted module, avoiding a reverse import
  of `tools.bot.run` while `python -m tools.bot.run` executes as `__main__`.
- Lowered the `tools/bot/run.py` maintainability baseline from 4612 to 4546 lines and added the new
  wait runtime module to the CODEMAP Bot / scenarios row.

## Proof

- `python -m py_compile tools/bot/run.py tools/bot/wait_runtime.py`
- `make test-py`
- `make bot`
- `.venv/bin/python tools/validate_codemap.py`
- `make maintainability`
- `make ci`

## Deferred

- Splitting action dispatch, state ingestion, co-op orchestration, and replay helpers out of
  `tools/bot/run.py` remains future paydown.

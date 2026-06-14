# v147 Spec: Bot Wait Runtime Split

Status: Complete
Date: 2026-06-14
Codename: `bot-wait-runtime-split`

## Purpose

Move the Python protocol bot's WebSocket wait/pump helpers out of `tools/bot/run.py` into a focused
runtime module while preserving scenario behavior and the existing public helper names used by the
bot runner and extracted movement module.

## Non-goals

- No scenario JSON format changes.
- No new bot step types, assertions, or wait semantics.
- No protocol, server, shared-rule, replay, or Godot client changes.
- No broader split of action dispatch, state ingestion, co-op orchestration, or CLI entrypoint code.

## Acceptance Criteria

- A new `tools/bot/wait_runtime.py` module owns wait/pump helpers for accepts, rejects, events,
  progression/cooldown waits, level/teleporter waits, player-position waits, and message pumping.
- `tools/bot/run.py` keeps compatibility wrappers for moved helper names so existing internal call
  sites and previously extracted modules continue to work.
- `tools/bot/run.py` shrinks and its maintainability baseline is lowered.
- Python bot unit checks and the full protocol bot remain green.

## Scope And Likely Files

- `tools/bot/run.py` - replace wait/pump helper bodies with wrappers.
- `tools/bot/wait_runtime.py` - new focused wait/pump helper module.
- `.maintainability/file-size-baseline.tsv` - lower `run.py` baseline.
- `docs/CODEMAP.md` and lifecycle docs.

## Test And Bot Proof

- `python -m py_compile tools/bot/run.py tools/bot/wait_runtime.py`
- `make test-py`
- `make bot`
- `make maintainability`
- `make ci`

## Open Questions And Risks

- No blocking product/design questions.
- Risk: extracted wait helpers need access to assertion and event-match helpers from `run.py`.
  Mitigation: keep wrappers in `run.py` and pass helper globals into the extracted module instead of
  importing `tools.bot.run` back from the helper.

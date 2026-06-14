# v148 Spec: Bot State Ingest Split

Status: Complete
Date: 2026-06-14
Codename: `bot-state-ingest-split`

## Purpose

Move the Python protocol bot's WebSocket state ingestion and small runtime state mutation helpers out
of `tools/bot/run.py` into a focused module while preserving scenario behavior and replay proof.

## Non-goals

- No scenario JSON format changes.
- No state delta, snapshot, protocol, server, shared-rule, replay, or Godot client changes.
- No new bot assertions or action behavior changes.
- No broader split of action dispatch, co-op orchestration, or CLI entrypoint code.

## Acceptance Criteria

- A new `tools/bot/state_ingest.py` module owns session snapshot ingestion, state delta ingestion,
  discovered-teleporter parsing, inventory/stash/hotbar upserts/removals, cooldown decay, active
  level clearing, initial-position tracking, and runtime distance tracking.
- `tools/bot/run.py` keeps compatibility wrappers for moved helper names so existing internal call
  sites and previously extracted modules continue to work.
- `tools/bot/run.py` shrinks and its maintainability baseline is lowered.
- Python bot unit checks and the full protocol bot remain green.

## Scope And Likely Files

- `tools/bot/run.py` - replace state ingestion/helper bodies with wrappers.
- `tools/bot/state_ingest.py` - new focused state ingestion module.
- `.maintainability/file-size-baseline.tsv` - lower `run.py` baseline.
- `docs/CODEMAP.md` and lifecycle docs.

## Test And Bot Proof

- `python -m py_compile tools/bot/run.py tools/bot/state_ingest.py`
- `make test-py`
- `make bot`
- `make maintainability`
- `make ci`

## Open Questions And Risks

- No blocking product/design questions.
- Risk: extracted ingestion still depends on finder/logging helpers from `run.py`. Mitigation: keep
  wrappers in `run.py` and pass helper globals into the extracted module instead of importing
  `tools.bot.run` back from the helper.

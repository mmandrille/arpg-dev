# v146 Spec: Bot Movement Runtime Split

Status: Complete
Date: 2026-06-14
Codename: `bot-movement-runtime-split`

## Purpose

Move the Python protocol bot's movement/pathing runtime helpers out of `tools/bot/run.py` into a
focused module while preserving scenario behavior and the existing public helper function names used
inside the bot runner.

## Non-goals

- No scenario JSON format changes.
- No new bot step types or behavior changes.
- No protocol, server, shared-rule, replay, or Godot client changes.
- No broader split of action dispatch, state ingestion, co-op orchestration, or CLI entrypoint code.

## Acceptance Criteria

- A new `tools/bot/movement_runtime.py` module owns movement helpers such as walking toward a
  target, move-to-position, in-range movement, range candidate calculation, and player movement
  accept/reject waiting.
- `tools/bot/run.py` keeps compatibility wrappers for the moved helper names so existing internal
  call sites continue to work during the staged split.
- `tools/bot/run.py` shrinks and its maintainability baseline is lowered.
- Python bot unit checks and the full protocol bot remain green.

## Scope And Likely Files

- `tools/bot/run.py` - replace movement helper bodies with wrappers.
- `tools/bot/movement_runtime.py` - new focused movement helper module.
- `.maintainability/file-size-baseline.tsv` - lower `run.py` baseline.
- `docs/CODEMAP.md` and lifecycle docs.

## Test And Bot Proof

- `python -m py_compile tools/bot/run.py tools/bot/movement_runtime.py`
- `make test-py`
- `make bot`
- `make maintainability`
- `make ci`

## Open Questions And Risks

- No blocking product/design questions.
- Risk: extracted helpers still need access to `run.py` local helper bindings while `python -m
  tools.bot.run` executes as `__main__`. Mitigation: keep wrappers in `run.py` and pass helper
  globals into the extracted module instead of importing `tools.bot.run` back from the helper.

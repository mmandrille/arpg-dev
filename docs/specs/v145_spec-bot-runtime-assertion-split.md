# v145 Spec: Bot Runtime Assertion Split

Status: Complete
Date: 2026-06-13
Codename: `bot-runtime-assertion-split`

## Purpose

Move the Python protocol bot's long snapshot/runtime assertion dispatch chains out of
`tools/bot/run.py` into a focused assertion module while preserving the existing public helper
functions, scenario semantics, and bot proof behavior.

## Non-goals

- No scenario JSON format changes.
- No new assertion types or assertion behavior changes.
- No server, protocol, gameplay, shared-rule, or Godot client changes.
- No broader split of action execution, movement helpers, co-op orchestration, or replay helpers.

## Acceptance Criteria

- `tools/bot/runtime_assertions.py` owns the extracted snapshot and runtime assertion dispatch.
- `tools.bot.run.run_assertions` and `tools.bot.run.run_runtime_assertions` remain callable as
  compatibility wrappers for existing tests and scenario execution.
- Existing helper modules such as `stash_assertions.py` and `unique_effect_assertions.py` remain the
  domain helper pattern; this slice does not inline their logic back into `run.py`.
- `tools/bot/run.py` shrinks and its maintainability baseline is lowered.
- Python bot unit checks and the full protocol bot remain green.

## Scope And Likely Files

- `tools/bot/run.py` - replace assertion dispatch bodies with thin wrappers.
- `tools/bot/runtime_assertions.py` - new focused assertion dispatch module.
- `tools/bot/test_item_assertions.py` or focused Python tests - keep compatibility coverage for
  importing `run_assertions` from `tools.bot.run`.
- `.maintainability/file-size-baseline.tsv` - lower `run.py` baseline.
- `docs/CODEMAP.md` and lifecycle docs.

## Test And Bot Proof

- `.venv/bin/pytest tools/bot/test_item_assertions.py tools/bot/test_stash_assertions.py`
- `make test-py`
- `make bot`
- `make maintainability`
- `make ci`

## Open Questions And Risks

- No blocking product/design questions.
- Risk: moving dispatch into a helper module accidentally changes import timing for `python -m
  tools.bot.run`. Mitigation: keep wrappers in `run.py` and pass existing helper globals into the
  extracted module instead of importing `tools.bot.run` back from the helper.

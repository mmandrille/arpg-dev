# v145 As-Built: Bot Runtime Assertion Split

Date: 2026-06-13
Spec: [`docs/specs/v145_spec-bot-runtime-assertion-split.md`](../specs/v145_spec-bot-runtime-assertion-split.md)
Plan: [`docs/plans/v145_2026-06-13-bot-runtime-assertion-split.md`](../plans/v145_2026-06-13-bot-runtime-assertion-split.md)

## What Shipped

- Moved Python bot snapshot and runtime assertion dispatch into
  `tools/bot/runtime_assertions.py`.
- Kept `tools.bot.run.run_assertions` and `tools.bot.run.run_runtime_assertions` as compatibility
  wrappers with the same public signatures used by tests and scenario execution.
- Passed existing helper bindings from `run.py` into the extracted module, avoiding a reverse import
  of `tools.bot.run` while `python -m tools.bot.run` executes as `__main__`.
- Lowered the `tools/bot/run.py` maintainability baseline from 5179 to 4768 lines and added the new
  assertion module to the CODEMAP Bot / scenarios row.

## Proof

- `python -m py_compile tools/bot/run.py tools/bot/runtime_assertions.py`
- `.venv/bin/pytest tools/bot/test_item_assertions.py tools/bot/test_stash_assertions.py`
- `make test-py`
- `make bot`
- `.venv/bin/python tools/validate_codemap.py`
- `make maintainability`
- `make ci`

## Deferred

- Splitting action execution, movement helpers, co-op orchestration, and replay helpers out of
  `tools/bot/run.py` remains future paydown.

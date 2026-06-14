# v149 As-Built: Bot Co-op Runtime Split

Date: 2026-06-14
Spec: [`docs/specs/v149_spec-bot-coop-runtime-split.md`](../specs/v149_spec-bot-coop-runtime-split.md)
Plan: [`docs/plans/v149_2026-06-14-bot-coop-runtime-split.md`](../plans/v149_2026-06-14-bot-coop-runtime-split.md)

## What Shipped

- Moved reusable Python protocol bot co-op runtime helpers into `tools/bot/coop_runtime.py`.
- Kept `tools.bot.run` compatibility wrappers for peer connect/close, peer message pumping, co-op
  waits, co-op intent sending, accept waiting, player position/entity selectors, and party-role
  assertion.
- Left scenario-specific co-op proof bodies in `tools/bot/run.py` to keep this slice narrow and
  behavior-preserving.
- Lowered the `tools/bot/run.py` maintainability baseline from 4288 to 4269 lines and added the new
  co-op runtime module to the CODEMAP Bot / scenarios row.

## Proof

- `python -m py_compile tools/bot/run.py tools/bot/coop_runtime.py`
- `make test-py`
- `make bot`
- `.venv/bin/python tools/validate_codemap.py`
- `make maintainability`
- `make ci`

## Deferred

- Splitting scenario-specific co-op proof bodies, action dispatch, and replay helpers out of
  `tools/bot/run.py` remains future paydown.

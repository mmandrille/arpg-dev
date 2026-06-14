# v149 Spec: Bot Co-op Runtime Split

Status: Complete
Date: 2026-06-14
Codename: `bot-coop-runtime-split`

## Purpose

Move the Python protocol bot's reusable co-op runtime helpers out of `tools/bot/run.py` into a
focused module while preserving all co-op scenario behavior and replay proof.

## Non-goals

- No scenario JSON format changes.
- No protocol, server, shared-rule, replay, or Godot client changes.
- No new co-op behavior or assertions.
- No extraction of the large scenario-specific co-op proof bodies in this slice.

## Acceptance Criteria

- A new `tools/bot/coop_runtime.py` module owns low-level co-op helper implementations for peer
  connect/close, peer message pumping, co-op waits, co-op intent sending, accept waiting, player
  position/entity selectors, and party-role assertion.
- `tools/bot/run.py` keeps compatibility wrappers for moved helper names so existing co-op scenario
  proof bodies continue to work.
- `tools/bot/run.py` shrinks and its maintainability baseline is lowered.
- Python bot unit checks and the full protocol bot remain green.

## Scope And Likely Files

- `tools/bot/run.py` - replace low-level co-op helper bodies with wrappers.
- `tools/bot/coop_runtime.py` - new focused co-op runtime helper module.
- `.maintainability/file-size-baseline.tsv` - lower `run.py` baseline.
- `docs/CODEMAP.md` and lifecycle docs.

## Test And Bot Proof

- `python -m py_compile tools/bot/run.py tools/bot/coop_runtime.py`
- `make test-py`
- `make bot`
- `make maintainability`
- `make ci`

## Open Questions And Risks

- No blocking product/design questions.
- Risk: extracted co-op helpers need access to WebSocket, protocol, logging, and state ingestion
  helpers from `run.py`. Mitigation: keep wrappers in `run.py` and pass helper globals into the
  extracted module instead of importing `tools.bot.run` back from the helper.

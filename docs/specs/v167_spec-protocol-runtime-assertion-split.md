# v167 Spec — Protocol Runtime Assertion Split

Status: Complete
Date: 2026-06-14
Codename: `protocol-runtime-assertion-split`

## Purpose

Move a coherent protocol bot runtime assertion domain out of `tools/bot/runtime_assertions.py` into
a focused helper module. Existing protocol bot scenario assertion syntax and behavior must remain
unchanged.

## Non-goals

- No new protocol bot assertion types.
- No scenario JSON, gameplay, server, protocol schema, or shared-rule changes.
- No full rewrite of `runtime_assertions.py`.

## Acceptance Criteria

- Runtime shop/stash economy assertions delegate to a focused helper module.
- Existing shop/stash runtime assertion tests continue to pass.
- `runtime_assertions.py` shrinks and remains importable without helper-global forwarding changes.
- `.venv/bin/pytest tools/bot/test_protocol.py`, `make maintainability`, and `make ci` pass.

## Scope and Likely Files

- Python bot: `tools/bot/runtime_assertions.py`
- New helper: `tools/bot/runtime_economy_assertions.py`
- Tests: existing `tools/bot/test_protocol.py`
- Docs: `PROGRESS.md`, `docs/as-built/v167_protocol-runtime-assertion-split.md`

## Test and Bot Proof

No new runtime scenario is needed because this preserves existing assertion semantics. Proof comes
from focused pytest coverage for runtime shop/stash assertions and full CI.

## Open Questions and Risks

- Risk: helper extraction can change assertion failure labels. Preserve existing helper calls and
  labels exactly where possible.

# v137 Spec: Bot Assertion Domain Split

Status: Complete
Date: 2026-06-13
Codename: `bot-assertion-domain-split`

## Purpose

Reduce pressure on the broad protocol bot runner by moving stash and unique-chest assertion helpers
into a focused bot-domain module. The player-visible behavior is unchanged; this slice improves the
test/tooling boundary so later unique, stash, and reward scenarios can add assertions without
growing `tools/bot/run.py` by default.

## Non-goals

- No gameplay, protocol, or shared-rule behavior changes.
- No broad bot runner rewrite or async action extraction.
- No new unique item behavior.

## Acceptance Criteria

- `tools/bot/run.py` imports stash/chest filtering and assertion helpers from a focused module.
- Existing protocol bot assertions for stash item count, stash gold, stash capacity, stash events,
  and unique chest item selection keep the same JSON shape.
- Focused Python tests cover the extracted helper module directly and the existing runtime/state
  assertion paths still pass.
- `make maintainability` and `make ci` pass before closeout.

## Scope and Likely Files

- `tools/bot/stash_assertions.py` - new focused helper module.
- `tools/bot/run.py` - import and dispatch through helper functions.
- `tools/bot/test_protocol.py` - existing assertion path coverage.
- `tools/bot/test_stash_assertions.py` - focused helper behavior coverage.
- `tools/bot/test_item_assertions.py` - regression coverage for opt-in rolled item display-name suffix checks.
- Bot/CI scripts and existing scenario JSON - deterministic closeout repairs for debug-gated unique chest and bishop client bot coverage.
- `docs/plans/v137_2026-06-13-bot-assertion-domain-split.md`
- `docs/as-built/v137_bot-assertion-domain-split.md`
- `PROGRESS.md`

## Test and Bot Proof

- `.venv/bin/pytest tools/bot/test_protocol.py tools/bot/test_stash_assertions.py tools/bot/test_item_assertions.py -q`
- `make maintainability`
- `make ci`

No new bot scenario is needed because the slice keeps existing gameplay behavior; existing stash,
unique-chest, bishop, and client bot scenarios continue to exercise the same player-facing paths.

## Open Questions and Risks

- Risk: circular imports from the new helper back into `run.py`. Mitigation: keep the helper pure
  and pass dependencies such as the count comparator as callables from `run.py`.

# v151 Spec — Extraction Independence Gate

Status: Complete
Date: 2026-06-14
Codename: `extraction-independence-gate`

## Purpose

Stop counting namespace-relocation as architectural decoupling. A newly extracted module should be
importable and unit-testable without importing the file it came from or receiving that file's full
`globals()` namespace. This slice turns that rule into a lightweight CI-backed ratchet and records
the policy so future structural slices have a higher proof bar than line-count shrinkage.

## Non-goals

- Do not continue the dedicated `tools/bot/run.py` split campaign in this slice.
- Do not introduce a typed `BotContext`; that is a larger one-shot refactor only worth doing if the
  bot runtime needs more modular work later.
- Do not rewrite v145-v149 helper bindings. The existing `helpers=globals()` sites are legacy debt
  and are baselined, not normalized.
- Do not split `server/internal/game/game_test.go`; use touch-to-shrink opportunistically in later
  gameplay slices.

## Acceptance Criteria

- `make maintainability` fails when a tracked source/tool file adds a new `helpers=globals()` helper
  injection above the current baseline.
- `make maintainability` fails when a helper-global injection baseline is stale after removals, so
  reductions must be locked in.
- A focused pytest fixture proves pass, growth-failure, stale-baseline failure, and unbaselined-file
  failure behavior for the new gate.
- `CLAUDE.md` states that extracted modules only count as decoupled when they can be imported and
  unit-tested without importing the source file or receiving its whole namespace.
- `PROGRESS.md` records that the dedicated `run.py` split campaign is frozen unless a future slice
  performs the real typed-context refactor.

## Scope and Files Likely Touched

- `scripts/check-extraction-coupling-ratchet.py` — new CI ratchet for helper-global namespace coupling.
- `.maintainability/extraction-coupling-baseline.tsv` — baseline current legacy coupling sites.
- `make/ci.mk` — wire the new gate into `make maintainability`.
- `tools/test_extraction_coupling_ratchet.py` — focused unit tests for the gate.
- `CLAUDE.md` — policy for extraction independence and importable/testable modules.
- `PROGRESS.md`, `docs/as-built/v151_extraction-independence-gate.md` — lifecycle closeout.

## Test and Bot Proof

- `python3 scripts/check-extraction-coupling-ratchet.py`
- `.venv/bin/pytest tools/test_extraction_coupling_ratchet.py -q`
- `make maintainability`
- `make ci`

No gameplay, protocol, world, inventory, movement, or replay behavior changes are in scope, so no
bot scenario is required.

## Open Questions and Risks

- Risk: this first gate catches the known laundering mechanism directly (`helpers=globals()`), not
  every possible bad extraction pattern. Mitigation: the policy requires focused import/unit proof,
  and future tooling can add more patterns when new abuse appears.

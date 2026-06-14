# v151 As-built — Extraction Independence Gate

## What shipped

- Added `scripts/check-extraction-coupling-ratchet.py`, a maintainability gate that counts
  `helpers=globals()` helper-global injection sites in tracked source/tool files.
- Added `.maintainability/extraction-coupling-baseline.tsv` with the current legacy
  `tools/bot/run.py` baseline of 43 occurrences.
- Wired the gate into `make maintainability` after the existing file-size ratchet.
- Added focused pytest coverage for pass, count-growth failure, stale-baseline failure, and
  unbaselined-file failure behavior.
- Updated `CLAUDE.md` so extracted modules only count as decoupled when they can be imported and
  unit-tested without importing the source file or receiving its whole namespace.

## Key decisions

- Existing `helpers=globals()` wrappers are grandfathered debt, not normalized in this slice.
- The dedicated `run.py` split campaign is frozen. Future `run.py` modularization should be a
  typed `BotContext` refactor if the bot runtime is worth more structural work.
- This gate targets the known laundering pattern directly. Broader extraction quality still relies
  on spec/plan review requiring direct import and focused tests for new modules.

## Verification

- `python3 scripts/check-extraction-coupling-ratchet.py`
- `.venv/bin/pytest tools/test_extraction_coupling_ratchet.py -q`
- `make maintainability`
- `make ci`

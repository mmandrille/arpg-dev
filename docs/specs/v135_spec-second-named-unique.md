# v135 Spec: Second Named Unique

Status: Complete
Date: 2026-06-13
Codename: `second-named-unique`

## Purpose

Add a second enabled named unique item to the shared catalog so the deterministic purple town unique
chest exposes more than one hand-authored unique package. The item should use an existing live unique
effect and an existing base template, proving that the named unique path supports catalog expansion.

## Non-goals

- No new unique effect mechanics, combat behavior, drop odds, or market restrictions.
- No new item templates, item presentation assets, or tooltip layout changes.
- No changes to natural random unique generation.
- No broad validation rewrite.

## Acceptance Criteria

- `shared/rules/unique_items.v0.json` defines a second enabled ready named unique with fixed stats,
  fixed effect ids, behavior hook text, and a minimum level.
- The new named unique uses an existing live effect compatible with its base template type.
- Go rules can build deterministic payloads for both enabled named uniques.
- The purple town unique chest includes both enabled named uniques in addition to effect-coverage
  rows.
- Shared validation and focused rule tests remain green.

## Likely Files

- `shared/rules/unique_items.v0.json`
- `server/internal/game/unique_chest_test.go`
- `tools/test_validate_unique_items.py`
- `PROGRESS.md`
- `docs/as-built/v135_second-named-unique.md`

## Test And Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestNamedUnique|TestUniqueTestChest'`
- `.venv/bin/python -m pytest tools/test_validate_unique_items.py -q`
- `make maintainability`
- `make ci`

No new visual bot scenario is required. Existing `purple_town_unique_chest` and client unique tooltip
coverage should prove the expanded deterministic chest remains usable.

## Open Questions And Risks

- Keep the second entry conservative: use a current base template and live effect rather than
  introducing new mechanics.

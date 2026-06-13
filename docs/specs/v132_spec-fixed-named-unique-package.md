# v132 Spec: Fixed Named Unique Package

Status: In progress
Date: 2026-06-13
Codename: `fixed-named-unique-package`

## Purpose

Turn the ready `embercall_blade` unique catalog entry from metadata into a deterministic fixed item
package. The purple town unique chest should now grant the named Embercall Blade in addition to the
effect-coverage rows, so testers can inspect at least one hand-authored named unique item.

## Non-goals

- No natural drop-rate changes.
- No mystery-seller or market unique restrictions.
- No new unique effects or effect mechanics.
- No new client inspection UI or production unique art.
- No full catalog expansion beyond the existing `embercall_blade` entry.

## Acceptance Criteria

- `shared/rules/unique_items.v0.json` defines fixed stats and fixed effect ids for enabled ready
  named uniques.
- Shared validation rejects named uniques whose fixed effect ids are unknown, disabled, duplicated,
  or incompatible with the base template item type.
- Go rules load the named unique catalog and can build a deterministic `ItemRollPayload` for
  `embercall_blade`.
- Opening `town_unique_chest` grants the named Embercall Blade plus the existing enabled-effect
  coverage rows.
- The protocol bot scenario `purple_town_unique_chest` asserts both full effect coverage and
  presence of the named unique.

## Likely Files

- `shared/rules/unique_items.v0.schema.json`
- `shared/rules/unique_items.v0.json`
- `tools/validate_shared.py`
- `server/internal/game/rules.go`
- `server/internal/game/unique_chest.go`
- `server/internal/game/unique_chest_test.go`
- `tools/bot/unique_effect_assertions.py`
- `tools/bot/scenarios/61_purple_town_unique_chest.json`
- `tools/bot/test_protocol.py`
- `PROGRESS.md`
- `docs/as-built/v132_fixed-named-unique-package.md`

## Test And Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestNamedUnique|TestUniqueTestChest|TestLoadRules'`
- `.venv/bin/python -m pytest tools/bot/test_protocol.py -q`
- `make bot scenario=purple_town_unique_chest`
- `make maintainability`
- `make ci`

## Open Questions And Risks

- None blocking. The named unique uses the existing `everburning_wound` live effect because it is
  compatible with `cave_blade` and matches the Embercall concept.

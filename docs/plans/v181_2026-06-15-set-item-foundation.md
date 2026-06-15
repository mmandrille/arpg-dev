# v181 Plan: Set Item Foundation

Status: Ready for implementation
Goal: Add a first five-piece green set-item package and prove equipped-piece bonuses through the debug chest.
Architecture: Set items are data-driven fixed rolled payloads from `set_items.v0.json`. The server counts equipped set pieces from durable payload metadata and adds derived bonuses during stat aggregation. The existing unique chest test container exposes the pieces for testing; client work is presentation-only rarity coloring.
Tech stack: shared JSON/schema, Go sim, Godot client tests, lifecycle docs.

## Baseline and shortcut decision

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.json` | Add `set` rarity metadata. |
| Add | `shared/rules/set_items.v0.json` | Define the first five-piece set and bonuses. |
| Add | `shared/rules/set_items.v0.schema.json` | Validate set catalog shape. |
| Modify | `server/internal/game/rules.go` | Load and validate set catalog. |
| Add | `server/internal/game/set_items.go` | Build set payloads and aggregate equipped bonuses. |
| Modify | `server/internal/game/types.go` | Persist set metadata in rolled payloads. |
| Modify | `server/internal/game/unique_chest.go` | Include set items in debug chest. |
| Modify | `server/internal/game/unique_chest_test.go` | Cover chest inclusion and set bonuses. |
| Modify | `client/scripts/main.gd`, `stash_panel.gd`, `shop_panel.gd` | Add green set rarity color. |
| Modify | `client/tests/test_item_visuals.gd`, `test_shop_panel.gd` | Assert green set presentation. |
| Add | `docs/as-built/v181_set-item-foundation.md` | As-built notes. |
| Modify | `PROGRESS.md` | Mark v181 complete. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd` stays within ratchet by a tiny color-map edit.
- [x] `server/internal/game/game_test.go` not touched.
- [x] `tools/bot/run.py` not touched.
- [x] Other over-limit file: `server/internal/game/unique_chest_test.go` receives focused assertions only.

Decision:
- [x] Add focused `set_items.go` for new server domain.

## Task 1 - Shared set rules

- [x] Add `set` rarity.
- [x] Add `set_items.v0.json` and schema.
- [x] Validate shared data.

## Task 2 - Server set item behavior

- [x] Load/validate set catalog.
- [x] Add durable set metadata to rolled payloads.
- [x] Include set items in the debug chest.
- [x] Add equipped set bonus aggregation.

## Task 3 - Client presentation

- [x] Add green set rarity colors to loot, stash, and shop surfaces.
- [x] Add focused client assertions.

## Task 4 - Lifecycle and CI

- [x] Update docs and `PROGRESS.md`.
- [x] Run focused verification and `make ci`.

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestUnique|TestSet'`
- [x] `make client-unit`
- [x] `make ci`

## Deferred scope

- Random set drops, drop weighting, dedicated set chest tabs, richer set tooltips, set collection tracking, and production set art.

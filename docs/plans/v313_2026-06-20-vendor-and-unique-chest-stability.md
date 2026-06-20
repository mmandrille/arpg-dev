# v313 Plan - Vendor And Unique Chest Stability

Status: Complete
Goal: Apply the v312 tooltip stability pattern to shop-style panels and backfill stale persisted unique test chest state.
Architecture: Keep UI stability in shared client render/tooltip guards. Keep unique test chest catalog repair server-side so all clients receive corrected chest contents.
Tech stack: Godot 4 GDScript client panel tests and Go server unit tests.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/shop_panel.gd` | Skip unchanged vendor/mystery renders |
| Modify | `client/scripts/main.gd` | Keep account-stash refreshes from overwriting open unique chest contents |
| Modify | `client/scripts/inventory_render_guard.gd` | Include shop offer state in stable render keys |
| Modify | `client/tests/test_stash_panel.gd` | Prove unique chest counts survive account-stash refreshes |
| Create | `client/tests/test_shop_tooltip_stability.gd` | Prove stable vendor/mystery slot refresh and tooltip mouse-ignore |
| Modify | `scripts/client_smoke.sh` | Register the focused shop tooltip stability test |
| Modify | `server/internal/game/unique_chest.go` | Backfill missing current catalog entries into persisted unique chest state |
| Modify | `server/internal/game/unique_chest_test.go` | Prove stale unique chest state is repaired |
| Create | `docs/as-built/v313_vendor-and-unique-chest-stability.md` | Completion proof |

## Tasks

- [x] Guard unchanged shop renders for town vendor and mystery seller.
- [x] Prove vendor/mystery item tooltip trees ignore mouse input.
- [x] Guard unique chest panels against account-stash refresh overwrite.
- [x] Add server backfill for persisted unique test chest catalog gaps.
- [x] Prove repeat unique chest opens and stale state repair.
- [x] Run final client and maintainability gates.

## Verification

- [x] `cd server && go test ./internal/game -run 'TestUniqueTestChest(OpensContainerAndTakesSelectedItem|BackfillsPersistedCatalogGaps|RepeatActivationReopensRemainingItems)' -count=1`
- [x] `godot --headless --path client --script res://tests/test_shop_tooltip_stability.gd`
- [x] `godot --headless --path client --script res://tests/test_shop_panel.gd`
- [x] `godot --headless --path client --script res://tests/test_stash_panel.gd`
- [x] `cd server && go test ./internal/game -run UniqueTestChest -count=1`
- [x] `make client-unit`
- [x] `make maintainability`

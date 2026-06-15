# v196 Plan: Item Level Progression

## Spec

`docs/specs/v196_spec-item-level-progression.md`

## Tasks

1. Add `ItemLevel` to `ItemRollPayload` and propagate it through clone/load helpers.
2. Populate generated template item levels from source depth, clamped to at least 1.
3. Populate named unique and set package item levels from effective minimum level.
4. Surface `item_level` on entity, inventory, stash, shop offer, and appraisal views.
5. Update v8 protocol schemas for the new optional rolled-item field.
6. Add bot assertion support and a deterministic v196 scenario.
7. Add focused server tests for depth mapping and protocol propagation.
8. Update as-built docs and `PROGRESS.md`, then run verification and commit.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'ItemLevel|RolledTemplateLootTransfersToInventory' -count=1`
- `make bot scenario=85_item_level_progression.json`
- `make ci`

## Notes

- No branch changes.
- No client visual work, so Godot plugin adoption is rejected for this slice.

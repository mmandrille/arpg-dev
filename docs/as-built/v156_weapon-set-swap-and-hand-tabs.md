# v156 As-Built - Weapon Set Swap and Hand Tabs

Spec: [`docs/specs/v156_spec-weapon-set-swap-and-hand-tabs.md`](../specs/v156_spec-weapon-set-swap-and-hand-tabs.md)
Plan: [`docs/plans/v156_2026-06-14-weapon-set-swap-and-hand-tabs.md`](../plans/v156_2026-06-14-weapon-set-swap-and-hand-tabs.md)

## What shipped

- Added two authoritative weapon sets with active-set compatibility through existing
  `equipped.main_hand` / `equipped.off_hand` fields.
- Added `swap_weapon_set_intent`, optional `weapon_set` targeting for hand equip/unequip, and v8
  snapshot/delta fields for `active_weapon_set` and `weapon_sets`.
- Added server tests for inactive-set equip, active swap behavior, per-set hand blocking, and
  active ranged attack mode after swapping to a bow.
- Added inventory paper-doll hand tabs for viewing/equipping set I and set II.
- Bound `R` in the Godot client to the authoritative weapon-set swap intent.
- Added protocol bot proof:
  - `make bot scenario=66_weapon_set_swap_and_tabs.json`

## Key decisions

- Rejected inventory UI plugins for this slice and extended the existing panel directly.
- Kept non-hand equipment shared across both weapon sets.
- Kept old `equipped` hand fields as the active weapon set so existing combat, visuals, and tests
  continue to consume one active loadout.

## Deferred

- Durable storage schema for two weapon sets beyond in-session state.
- More polished hand-tab art and tooltips.
- Settings/rebinding for the swap key.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestWeaponSet|TestFullEquipment|TestEquipped' -count=1`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make client-unit`
- `make bot scenario=66_weapon_set_swap_and_tabs.json`
- `make maintainability`
- `make ci`

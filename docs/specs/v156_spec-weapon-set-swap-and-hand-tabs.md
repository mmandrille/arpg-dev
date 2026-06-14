# v156 Spec - Weapon Set Swap and Hand Tabs

## Purpose

Add a Diablo-style second weapon configuration so players can equip two hand sets, swap the active
set with `R`, and inspect/equip either hand set from the inventory paper doll.

This combines the selected `$autoloop` queue items 9+10:

- second equipable weapon set
- inventory silhouette/tabs to manually view both hand configurations

## Player Experience

- The character owns two weapon sets: set 1 and set 2.
- Only one set is active for combat, stats, visuals, and skill weapon requirements.
- Pressing `R` swaps active set 1 <-> set 2.
- The inventory paper doll exposes two hand tabs near the main/off hand slots.
- Selecting a hand tab shows that set's main/off hand items.
- Dragging or double-click equipping a weapon while a hand tab is selected equips into that viewed
  set.
- Non-hand equipment remains shared across both weapon sets.

## Functional Requirements

### Server authority

- Server state stores:
  - active weapon set index: `0` or `1`
  - hand equipment for both weapon sets
  - existing non-hand equipment as shared equipment
- Snapshot and delta payloads expose enough state for clients to render both hand sets and the active
  set.
- Combat, damage type, attack speed, skill weapon checks, equipped unique effects, item-skill stat
  bonuses, paper-doll visuals, and shop comparisons use only the active weapon set's hands.
- Equipping to `main_hand` / `off_hand` without an explicit set targets the active set to preserve
  existing bot/client behavior.
- Equipping to an inactive hand set validates slot, class, handedness, requirements, hand blocking,
  and capacity with the same rules as active set equips.
- Unequipping a hand slot can target the active set by default and a specified set when provided.
- Swapping sets emits an authoritative event and equipment updates so clients/replay converge.
- Existing non-hand slots and inventory capacity rules keep current behavior.

### Protocol

- Add `swap_weapon_set_intent`.
- Extend `equip_intent` and `unequip_intent` with optional `weapon_set`.
- Extend snapshots/deltas with active weapon-set and per-set hand equipment fields.
- Keep older `equipped.main_hand` / `equipped.off_hand` populated with the active set for current
  consumers.

### Client

- Bind `R` to send `swap_weapon_set_intent` when gameplay input is active.
- Inventory panel shows hand tabs `I` and `II` near the main/off hand slots.
- The selected hand tab controls which hand-set items are shown and where hand equipment drops go.
- The active hand tab is visually distinct from a merely viewed inactive tab.
- Equipment visuals use the active hand set from authoritative state.

## Non-Goals

- More than two weapon sets.
- Separate armor/jewelry sets.
- New weapon types or balance changes.
- Local-only speculative swapping.
- Rebinding the `R` key in settings.
- Full Diablo II artwork replication; the UI should use the existing inventory panel style.

## Acceptance Criteria

- Go tests prove active/inactive hand equips, swap behavior, hand blocking, default active-set
  compatibility, and active-set combat damage.
- Protocol schemas validate `swap_weapon_set_intent`, optional hand-set targeting, and new snapshot
  state.
- Client unit tests prove `R` emits the swap intent and inventory hand tabs route equip intents to
  the viewed set.
- Protocol bot scenario equips two weapon sets, swaps with the new intent, and observes combat from
  the active set.
- `make validate-shared`, focused Go tests, `make client-unit`, the new bot scenario,
  `make maintainability`, and `make ci` pass.

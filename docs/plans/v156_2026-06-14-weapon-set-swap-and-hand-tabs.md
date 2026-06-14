# v156 Plan - Weapon Set Swap and Hand Tabs

Spec: [`docs/specs/v156_spec-weapon-set-swap-and-hand-tabs.md`](../specs/v156_spec-weapon-set-swap-and-hand-tabs.md)

## Plugin / shortcut adoption checklist

- **Inventory UI plugins:** reject for this slice. GLoot/Godot-Inventory remain useful references,
  but this change is a focused extension of the existing authoritative inventory panel. Adopting a
  plugin would add migration work and risks duplicating item/equip logic client-side.
- **Borrowed pattern:** Diablo-style hand tabs only: two small tabs near hand slots, selected tab
  controls the visible hand set, active tab is highlighted separately.
- **Authority boundary:** server owns both weapon sets and active set; client sends intents only.

## Implementation Tasks

### 1. Server weapon-set model

Files:

- `server/internal/game/sim.go`
- `server/internal/game/handlers.go`
- `server/internal/game/sim_load.go`
- `server/internal/game/sim_players.go`
- `server/internal/game/types.go`
- new focused test file, likely `server/internal/game/weapon_sets_test.go`

Steps:

- Add constants for two weapon sets and a `weaponSetHands` structure.
- Keep `s.equipped` as the active/shared compatibility map, but back hand slots with a two-set
  authoritative store.
- Add helpers:
  - active hand lookup
  - set-targeted hand lookup/update
  - active equipped snapshot map
  - weapon-set snapshot view
- Update equip/unequip paths so hand slots can target optional `weapon_set`, defaulting to active.
- Preserve existing hand-clearing and two-handed/offhand validation inside the selected set.
- Add `swap_weapon_set_intent` handling that flips active set and emits updates/events.
- Update combat/stat/unique-effect helpers that read hand slots to use active hand set.
- Keep non-hand equipment unchanged.

### 2. Protocol and validation

Files:

- `shared/protocol/messages.v*.schema.json`
- `shared/protocol/session_snapshot.v*.schema.json`
- `shared/protocol/state_delta.v*.schema.json`
- `shared/protocol/examples/*.json`
- `tools/validate_shared.py` if cross-checks need awareness

Steps:

- Add `swap_weapon_set_intent` message payload.
- Add optional `weapon_set` integer (`0` or `1`) to equip/unequip payloads.
- Add snapshot fields:
  - `active_weapon_set`
  - `weapon_sets` with `main_hand` and `off_hand` IDs per set
- Add delta fields for active weapon-set changes and hand-set updates.
- Keep `equipped.main_hand` and `equipped.off_hand` as active-set compatibility values.

### 3. Client input and inventory tabs

Files:

- `client/scripts/main.gd`
- `client/scripts/inventory_panel.gd`
- `client/tests/test_delta_apply.gd`
- `client/tests/test_shop_panel.gd` or a new focused inventory-panel test

Steps:

- Store `active_weapon_set` and `weapon_sets` from snapshots/deltas.
- Send `swap_weapon_set_intent` on `R` during gameplay when input is not blocked.
- Pass weapon-set state into `InventoryPanel.set_inventory_state`.
- Add two compact hand tabs near the hand slots.
- Render main/off hand slots from selected hand tab.
- Include `weapon_set` on hand-slot equip/unequip intents when the selected tab is not implicit
  active state.
- Keep armor/jewelry rendering unchanged.

### 4. Bot proof

Files:

- `tools/bot/run.py` or extracted bot runtime module if needed
- `tools/bot/runtime_assertions.py`
- `tools/bot/scenarios/66_weapon_set_swap_and_tabs.json`
- `tools/bot/test_protocol.py`

Steps:

- Add bot action for `swap_weapon_set`.
- Add assertions for active weapon set and per-set equipped hand definitions.
- Scenario:
  - load a lab with two one-handed weapons or weapon + bow
  - equip set 1
  - equip set 2 via `weapon_set`
  - swap
  - assert active set and combat event comes from active set

### 5. Docs and finish

Files:

- `docs/as-built/v156_weapon-set-swap-and-hand-tabs.md`
- `PROGRESS.md`

Steps:

- Record shipped behavior and deferred UI polish.
- Run:
  - `make validate-shared`
  - focused Go tests for weapon sets
  - `make client-unit`
  - `make bot scenario=66_weapon_set_swap_and_tabs.json`
  - `make maintainability`
  - `make ci`
- Commit as `feat: v156: weapon set swap and hand tabs`.

## Risks

- Many helper paths read `s.equipped[main_hand]`; missing one would cause active-set mismatches.
- Snapshot schema changes touch all protocol versions unless the repo has a local migration pattern
  for only the latest version.
- Capacity checks around inactive equipped items must not treat inactive set hands as bag items.

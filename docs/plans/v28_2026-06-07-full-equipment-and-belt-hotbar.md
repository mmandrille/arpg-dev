# v28 Plan — Full Equipment and Belt Hotbar

Status: Complete (`make ci` green on 2026-06-07)
Goal: Replace the single weapon slot with a full paper-doll equipment model, add server-persisted belt-gated hotbar layout, and prove every equipment category through deterministic drops, bots, replay, and client UI.
Architecture: The Go sim remains authoritative for slot compatibility, hand occupancy, hotbar capacity, hotbar use, item consumption, and persistence. Shared JSON rules define equipment slot metadata, handedness, rollable stat keys, template drops, and golden fixtures; Godot mirrors only presentation and sends intents. Hotbar assignment is character-scoped durable state, while each session freezes a start snapshot for deterministic replay. Direct bag use stays `use_intent { item_instance_id }`; hotbar key use becomes `use_hotbar_intent { slot_index }` so disabled slot rejection and character hotbar persistence have a concrete server-owned contract.
Tech stack: shared JSON schemas/rules/goldens, Go authoritative sim/store/realtime/replay/input decode, Postgres migration, Python protocol bot, Godot 4 GDScript inventory/hotbar UI, Godot client bot.

**Spec:** [`docs/specs/v28_spec-full-equipment-and-belt-hotbar.md`](../specs/v28_spec-full-equipment-and-belt-hotbar.md)

**Branch:** current checked-out branch `main` (no branch creation required)

---

## Spec review (2026-06-07)

| Check | Result |
|-------|--------|
| Baseline v28 after completed v27/v26 | OK |
| Branch | Resolved in chat: work on current `main`; ignore spec branch placeholder |
| Protocol contracts | OK with resolution: add `assign_hotbar_intent` and `use_hotbar_intent`; migrate `weapon` to `main_hand`/`off_hand` everywhere |
| Server authority | OK — client is presentation only; Go sim owns equip, hand rules, capacity, use, persistence |
| Determinism | OK — no wall clock/randomness outside seeded sim; equipment lab/golden must pin seed/order |
| Shared rules/goldens | OK — cross-language golden and validator updates required |
| Bot proof | OK — protocol scenario `19_full_equipment.json` and client scenario `10_full_equipment.json` required |
| Replay | OK — session-start equipped + hotbar snapshot and input replay must be updated |
| As-built drift | Current schemas, bots, Go sim, assets, and Godot UI are still `weapon`-centric; migration is explicit plan work |

Spec edits recommended before or during implementation:
- Change spec `Branch:` to current `main`.
- Replace disabled-slot wording that implies adding slot context to `use_intent` with `use_hotbar_intent { slot_index }`.
- Add `shared/protocol/examples/use_hotbar_intent.json` to the file list.

---

## Baseline and shortcut decision

v28 builds on v27 `hold-click-controls`, v26 `character-stats-and-leveling`, v23 rolled item templates, v22 character-scoped persistence, v16 consumable use, v15 item presentation metadata, v13 inventory UI, and v12 ranged weapon behavior.

The current as-built system has:
- Protocol v1 `equipped: { "weapon": ... }`
- `equip_intent` / `unequip_intent` restricted to `"weapon"`
- item/template slots using `"weapon"` for equippable weapons
- client-only `ConsumableBar` assignments
- store item rows with a single nullable `slot` plus `equipped`
- no durable hotbar table

Godot shortcut adoption checklist:

- **No new asset dependency:** placeholder category colors/icons through existing presentation metadata are enough for v28.

Protocol choice resolved during review:

- `assign_hotbar_intent { slot_index, item_instance_id | null }` mutates durable character hotbar assignment.
- `use_hotbar_intent { slot_index }` is the only hotbar-key server use path. It rejects `hotbar_slot_disabled` when `slot_index >= hotbar_capacity`, rejects empty/invalid/non-consumable cases, resolves the assigned item server-side, then consumes through the same authoritative logic as direct bag use.
- `use_intent { item_instance_id }` remains valid for direct bag item use and does not check hotbar capacity.

---

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/plans/v28_2026-06-07-full-equipment-and-belt-hotbar.md` | This plan |
| Modify | `docs/specs/v28_spec-full-equipment-and-belt-hotbar.md` | Align branch and `use_hotbar_intent` wording when implementation starts |
| Modify | `shared/rules/items.v0.schema.json` | Static item slot/hand metadata |
| Modify | `shared/rules/items.v0.json` | Migrate `weapon` slots to `main_hand`; add handedness/occupancy where needed |
| Modify | `shared/rules/item_templates.v0.schema.json` | Equipment slots, ring logical slot, handedness, rollable stat keys |
| Modify | `shared/rules/item_templates.v0.json` | One rollable template per equipment category |
| Modify | `shared/rules/treasure_classes.v0.schema.json` | Validate equipment lab TC entries if needed |
| Modify | `shared/rules/treasure_classes.v0.json` | Add `equipment_lab_tc_1` |
| Modify | `shared/rules/worlds.v0.json` | Add `equipment_lab` preset |
| Modify | `shared/rules/worlds.v0.schema.json` | Only if lab needs new world fields |
| Modify | `shared/assets/item_visuals.v0.schema.json` | Replace weapon-only visual slot enum |
| Modify | `shared/assets/item_visuals.v0.json` | Migrate sword/bow mappings to `main_hand`; add offhand if used |
| Modify | `shared/assets/item_presentations.v0.json` | Placeholder category icons/colors/names |
| Modify | `assets/manifests/assets.v0.json` | Slot metadata migration if existing equipment entries are slot-validated |
| Modify | `tools/assets/validate_assets.py` | Accept new slot ids and visual-slot agreement |
| Modify | `tools/assets/test_validate_assets.py` | Asset validator slot coverage |
| Modify | `shared/protocol/envelope.v1.schema.json` | Add `assign_hotbar_intent` and `use_hotbar_intent` types |
| Modify | `shared/protocol/messages.v1.schema.json` | Full slot enum; assign/use hotbar payloads |
| Modify | `shared/protocol/session_snapshot.v1.schema.json` | Full equipped map, `hotbar_capacity`, fixed 10-entry `hotbar` |
| Modify | `shared/protocol/state_delta.v1.schema.json` | Full equipped slot enum, `hotbar_update`, hotbar events if needed |
| Modify | `shared/protocol/examples/equip_intent.json` | Migrate to `main_hand` |
| Modify | `shared/protocol/examples/unequip_intent.json` | Migrate to `main_hand` |
| Create | `shared/protocol/examples/assign_hotbar_intent.json` | New hotbar assignment example |
| Create | `shared/protocol/examples/use_hotbar_intent.json` | New hotbar key-use example |
| Modify | `shared/protocol/examples/session_snapshot.json` | Full equipment + hotbar example |
| Modify | `shared/protocol/examples/state_delta.json` | Full equipment + hotbar deltas |
| Create | `shared/golden/full_equipment.v0.schema.json` | Golden fixture schema |
| Create | `shared/golden/full_equipment.json` | Equipment, hand occupancy, belt capacity, hotbar, shield roll fixture |
| Modify | `shared/golden/slice_outcome.v0.schema.json` | Migrate final equipped shape if still checked |
| Modify | `shared/golden/slice_outcome.json` | Migrate `weapon` to `main_hand` |
| Modify | `shared/golden/item_visual_resolution.json` | Migrate expected slot |
| Modify | `tools/validate_shared.py` | Slot/hand/hotbar/template/TC/golden drift checks |
| Create | `server/migrations/0006_character_hotbar_and_equipment.sql` | Durable hotbar and equipment migration |
| Modify | `server/internal/store/models.go` | Hotbar/equipment models |
| Modify | `server/internal/store/interfaces.go` | Hotbar repo methods |
| Modify | `server/internal/store/repos.go` | Postgres hotbar + multi-slot equipped persistence |
| Modify | `server/internal/store/store_test.go` | Hotbar and multi-slot persistence tests |
| Modify | `server/internal/game/rules.go` | Parse slots, handedness, occupancy, stat keys |
| Modify | `server/internal/game/types.go` | Equipped/hotbar views, intents, changes |
| Modify | `server/internal/game/sim.go` | Equip rules, primary attack lookup, hotbar capacity/use |
| Modify | `server/internal/game/game_test.go` | Golden and sim tests |
| Modify | `server/internal/inputdecode/inputdecode.go` | Decode assign/use hotbar intents |
| Modify | `server/internal/realtime/runner.go` | Persist hotbar and equipped slot mutations |
| Modify | `server/internal/replay/replay.go` | Session-start hotbar/equipped replay |
| Modify | `server/internal/http/session.go` | Load/create hotbar at session start |
| Modify | `server/internal/http/ws_test.go` | Snapshot/delta/reconnect assertions |
| Modify | `server/internal/replay/replay_test.go` | Replay migration assertions |
| Modify | `client/scripts/inventory_panel.gd` | Paper-doll layout and slot routing |
| Modify | `client/scripts/consumable_bar.gd` | Server-synced hotbar, capacity, disabled styling |
| Modify | `client/scripts/main.gd` | New intents, equipped map migration, snapshot/delta handling |
| Modify | `client/scripts/equipment_visuals.gd` | `main_hand` migration; optional offhand/bow mapping |
| Modify | `client/scripts/bot_controller.gd` | Client bot equipment/hotbar actions |
| Modify | `client/scripts/bot_scenario_runner.gd` | Client bot steps/assertions |
| Modify | `client/tests/test_golden.gd` | Full equipment golden checks |
| Create | `client/tests/test_inventory_equipment.gd` | Paper-doll/hotbar UI model tests if needed |
| Modify | `client/scripts/smoke.gd` | Migrate smoke weapon assumptions |
| Modify | `tools/bot/run.py` | Multi-slot equip and hotbar helpers/assertions |
| Modify | `tools/bot/test_protocol.py` | Helper/parser tests |
| Create | `tools/bot/scenarios/19_full_equipment.json` | Protocol end-to-end proof |
| Create | `tools/bot/scenarios/client/10_full_equipment.json` | Godot client UI proof |
| Modify | existing `tools/bot/scenarios/*.json` | Migrate `"weapon"` scenarios to `main_hand` |
| Modify | existing `tools/bot/scenarios/client/*.json` | Migrate weapon/hotbar assumptions |
| Modify | `PROGRESS.md` | Lifecycle update when slice ships |

---

## Task 1 — Shared slot, item, visual, and protocol contracts

Files:
- Modify: `shared/rules/items.v0.schema.json`
- Modify: `shared/rules/items.v0.json`
- Modify: `shared/rules/item_templates.v0.schema.json`
- Modify: `shared/rules/item_templates.v0.json`
- Modify: `shared/assets/item_visuals.v0.schema.json`
- Modify: `shared/assets/item_visuals.v0.json`
- Modify: `shared/assets/item_presentations.v0.json`
- Modify: `assets/manifests/assets.v0.json`
- Modify: `tools/assets/validate_assets.py`
- Modify: `tools/assets/test_validate_assets.py`
- Modify: `shared/protocol/envelope.v1.schema.json`
- Modify: `shared/protocol/messages.v1.schema.json`
- Modify: `shared/protocol/session_snapshot.v1.schema.json`
- Modify: `shared/protocol/state_delta.v1.schema.json`
- Modify/Create: protocol examples listed in the file map
- Create: `shared/golden/full_equipment.v0.schema.json`
- Create: `shared/golden/full_equipment.json`
- Modify: `shared/golden/slice_outcome.v0.schema.json`
- Modify: `shared/golden/slice_outcome.json`
- Modify: `shared/golden/item_visual_resolution.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Define the canonical equipment slot enum: `head`, `amulet`, `chest`, `gloves`, `belt`, `boots`, `ring_left`, `ring_right`, `main_hand`, `off_hand`, plus inventory row `slot: ""`; allow logical template slot `ring`.
- [x] Step 1.2: Replace all shared rule/schema references to equip slot `"weapon"` with `main_hand`, and add `handedness` / `occupies_hands` fields for hand equipment.
- [x] Step 1.3: Extend rolled stat vocabularies to include `armor`, `block_percent`, and `hotbar_slots`, while keeping damage stats authoritative and armor/block display-only in v28.
- [x] Step 1.4: Add the v28 template catalog: `cave_blade`, `cave_greatsword`, `cave_bow`, `cave_shield`, `cave_helm`, `cave_mail`, `cave_gloves`, `cave_belt`, `cave_boots`, `cave_ring`, and `cave_amulet`.
- [x] Step 1.5: Define `cave_belt` with `hotbar_slots` roll range `[3, 10]` and a base fallback value of at least `3`.
- [x] Step 1.6: Migrate item visual and asset slot validation from weapon-only to the full slot set; keep placeholder visuals simple.
- [x] Step 1.7: Add `assign_hotbar_intent` and `use_hotbar_intent` to protocol envelope/messages schemas.
- [x] Step 1.8: Define `assign_hotbar_intent { slot_index, item_instance_id }`, where `item_instance_id` may be null and `slot_index` is `0..9`.
- [x] Step 1.9: Define `use_hotbar_intent { slot_index }`, where `slot_index` is `0..9`; direct `use_intent { item_instance_id }` remains unchanged.
- [x] Step 1.10: Replace snapshot `equipped.required: ["weapon"]` with all paper-doll slots, each nullable decimal id.
- [x] Step 1.11: Add required snapshot fields `hotbar_capacity` and fixed 10-entry `hotbar` array with nullable item ids.
- [x] Step 1.12: Extend state deltas with `hotbar_update` and expanded `equipped_update` slot enum.
- [x] Step 1.13: Create `shared/golden/full_equipment.json` pinning 1H+shield, 2H/bow occupancy, belt capacity `10`, disabled assignment retention, re-enable behavior, and shield rolled stats.
- [x] Step 1.14: Update validator checks for slot migration, template category rules, TC coverage, hotbar capacity fixture drift, no lingering `"weapon"` in current protocol examples/goldens, and visual slot agreement.

```bash
make validate-shared
make validate-assets
```

---

## Task 2 — Equipment lab drops and world proof data

Files:
- Modify: `shared/rules/treasure_classes.v0.schema.json`
- Modify: `shared/rules/treasure_classes.v0.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `shared/rules/worlds.v0.schema.json` if needed
- Modify: `tools/validate_shared.py`
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add treasure class `equipment_lab_tc_1` with weighted entries covering every v28 equipment template and enough potion sources for hotbar proof.
- [x] Step 2.2: Add world preset `equipment_lab` with deterministic access to all required category drops. Prefer explicit preset loot/chests if needed to avoid RNG flakiness.
- [x] Step 2.3: If chests or existing TC mechanics cannot guarantee every category with a pinned seed, add only the minimal world/rule shape needed and validate it.
- [x] Step 2.4: Add validator coverage that `equipment_lab_tc_1` references known templates and covers every paper-doll category.
- [x] Step 2.5: Add Go rule tests proving the lab world and TC load deterministically.

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'EquipmentLab|TreasureClass|Rules'
```

---

## Task 3 — Store migration and character hotbar persistence

Files:
- Create: `server/migrations/0006_character_hotbar_and_equipment.sql`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/store/store_test.go`

- [x] Step 3.1: Add durable character hotbar storage keyed by `character_id` and `slot_index`, with nullable `item_instance_id`.
- [x] Step 3.2: Add immutable session-start hotbar storage keyed by `session_id` and `slot_index`.
- [x] Step 3.3: Ensure existing character item `slot` values can store all paper-doll slot ids; migrate any existing `"weapon"` rows to `main_hand`.
- [x] Step 3.4: Add repository models for hotbar rows and fixed 10-slot views.
- [x] Step 3.5: Add repo methods to get/create default hotbar, update/clear one slot, snapshot hotbar at session start, and load session-start hotbar.
- [x] Step 3.6: Preserve account/character ownership checks when assigning a hotbar item.
- [x] Step 3.7: Add store tests for default 10 slots, assignment, clearing, persistence across fresh sessions, session-start immutability, item deletion cleanup behavior, and migration from `weapon` to `main_hand`.

```bash
cd server && go test ./internal/store -run 'Hotbar|Equipment|CharacterItem|SessionStart'
```

Verification run:

```bash
cd server && go test ./internal/store
```

---

## Task 4 — Go sim equipment model and primary attack lookup

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 4.1: Replace `weaponSlot` and `equippedWeapon*` assumptions with a fixed equipped map over the paper-doll slot set.
- [x] Step 4.2: Implement slot compatibility for static items and rolled templates, including logical `ring` accepting `ring_left` or `ring_right`.
- [x] Step 4.3: Implement hand occupancy:
  - 1H main-hand weapon + offhand shield can coexist
  - 2H main-hand weapon/bow clears offhand when equipped
  - offhand equip rejects with `hands_blocked` when main hand holds a 2H item
  - bow uses existing ranged combat path and occupies both hands
- [x] Step 4.4: Preserve deterministic change ordering for atomic swaps/clears: inventory/equipped updates must be stable and replayable.
- [x] Step 4.5: Replace primary attack lookup with `main_hand` weapon resolution; do not let shields supply damage/reach.
- [x] Step 4.6: Keep rolled weapon damage precedence from v23, then static/template damage fallback, then unarmed fallback.
- [x] Step 4.7: Add Go tests for each paper-doll slot, ring alternatives, wrong slot rejection, 1H+shield, 2H clearing offhand, blocked offhand, bow ranged path, drop equipped item, unequip slot, and primary attack damage.
- [x] Step 4.8: Add golden-backed tests for `shared/golden/full_equipment.json`.

```bash
cd server && go test ./internal/game/... -run 'Equip|Equipment|Hand|Occup|Ring|Bow|Weapon|FullEquipment'
```

---

## Task 5 — Go sim hotbar capacity, assignment, and use

Files:
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/inputdecode/inputdecode.go`
- Modify: `shared/protocol/messages.v1.schema.json`

- [x] Step 5.1: Add `AssignHotbarIntent` and `UseHotbarIntent` to game input types and input decoding.
- [x] Step 5.2: Compute `hotbar_capacity` as `2` without belt, or `clamp(equipped_belt.hotbar_slots, 2, 10)` with base-stat fallback when the belt lacks a rolled value.
- [x] Step 5.3: Add hotbar state to `Snapshot` and state delta changes.
- [x] Step 5.4: Implement `assign_hotbar_intent`: reject malformed index, dead player, missing item, item not in bag, and non-consumable; allow assignment to disabled slots; emit/persist `hotbar_update`.
- [x] Step 5.5: Implement `use_hotbar_intent`: reject invalid index, disabled slot with `hotbar_slot_disabled`, empty slot, missing item, non-consumable, and dead player; on success, consume the assigned item.
- [x] Step 5.6: Keep direct `use_intent { item_instance_id }` valid for bag use independent of hotbar capacity.
- [x] Step 5.7: On successful consumable use, remove inventory row and clear every hotbar entry referencing the consumed instance.
- [x] Step 5.8: Emit authoritative `hotbar_update` deltas when assignment changes or consumed item references are cleared.
- [x] Step 5.9: Add tests for base capacity 2, rolled belt capacity, clamp max 10, disabled assignment retention, disabled hotbar-use rejection, direct bag use success, re-enable after belt re-equip, and hotbar clearing on consume/drop.

```bash
cd server && go test ./internal/game/... -run 'Hotbar|Belt|UseHotbar|UseConsumable|FullEquipment'
```

---

## Task 6 — Session bootstrap, realtime persistence, and replay

Files:
- Modify: `server/internal/http/session.go`
- Modify: `server/internal/realtime/runner.go`
- Modify: `server/internal/replay/replay.go`
- Modify: `server/internal/http/ws_test.go`
- Modify: `server/internal/replay/replay_test.go`
- Modify: `server/internal/http/auth_session_test.go` if fresh-session assertions belong there

- [x] Step 6.1: At fresh session creation, load durable character items/equipped slots and durable hotbar before creating session-start snapshots.
- [x] Step 6.2: Persist immutable session-start snapshots for equipped item locations and hotbar layout together.
- [x] Step 6.3: Feed fresh WebSocket snapshots from durable state and reconnect/replay from session-start snapshots plus recorded inputs.
- [x] Step 6.4: Persist `equipped_update`, inventory slot/location updates, `hotbar_update`, and consumed-item hotbar clears from realtime runner results.
- [x] Step 6.5: Ensure `/state`, WebSocket snapshot, replay timeline, and reconnect expose identical `equipped`, `hotbar_capacity`, and `hotbar` views.
- [x] Step 6.6: Add tests for same-session reconnect, deterministic replay, fresh-session persistence, selected-character isolation, and replay isolation from later character hotbar/equipment changes.

```bash
cd server && go test ./internal/http/... ./internal/replay/... -run 'Hotbar|Equipment|Replay|Reconnect|State|Session'
```

---

## Task 7 — Protocol bot migration and full equipment scenario

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`
- Create: `tools/bot/scenarios/19_full_equipment.json`
- Modify: existing `tools/bot/scenarios/01_vertical_slice.json` through `18_character_stats_and_leveling.json` as needed

- [x] Step 7.1: Migrate bot state parsing from `equipped.weapon` to the full equipped map.
- [x] Step 7.2: Migrate equip/unequip helper defaults from `"weapon"` to `main_hand`; allow slot override for all paper-doll slots.
- [x] Step 7.3: Add bot helpers/assertions for equipped slot id/def/template, empty slot, 1H+shield, offhand clear, hotbar assignment, capacity, disabled slot assigned, enabled/disabled hotbar use rejection/success, and fresh-session persistence.
- [x] Step 7.4: Add `assign_hotbar` and `use_hotbar_slot` scenario steps using the new protocol intents.
- [x] Step 7.5: Add `19_full_equipment.json` in `equipment_lab`: collect one item per category, equip all paper-doll slots, prove 1H+shield, swap to 2H, swap to bow, equip belt, assign potions to slots `0`, `2`, and `5`, unequip/re-equip belt, and use an enabled slot.
- [x] Step 7.6: Include `/state`, reconnect resume, replay verification, and fresh-session persistence checks for equipped and hotbar state.
- [x] Step 7.7: Update older protocol scenarios that mention slot `"weapon"` or assert `equipped.weapon`.
- [x] Step 7.8: Add or update Python unit tests for new scenario parser/assertion helpers.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
make bot
```

---

## Task 8 — Godot paper-doll inventory UI

Files:
- Modify: `client/scripts/inventory_panel.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Create: `client/tests/test_inventory_equipment.gd` if model extraction is needed

- [x] Step 8.1: Replace the single weapon slot UI with paper-doll slots laid out as head, rings/amulet, chest, hands, gloves/belt/boots, and bag grid.
- [x] Step 8.2: Store and render the full `equipped` map from snapshots and `equipped_update` deltas.
- [x] Step 8.3: Route double-click and drag/drop equip actions to `equip_intent { item_instance_id, slot }` for the target slot.
- [x] Step 8.4: Route drag equipped item to bag as `unequip_intent { slot }`; drag outside as existing `drop_intent`.
- [x] Step 8.5: Update client-side slot compatibility helpers for presentation-only affordances, including ring alternatives and hand slots. Server remains final authority.
- [x] Step 8.6: Extend tooltips to show display name, rarity, requirements, and rolled stats including `armor`, `block_percent`, and `hotbar_slots`.
- [x] Step 8.7: Keep panel input blocked by menu/pause/dead player as existing UI does.
- [x] Step 8.8: Add headless UI/model tests if the paper-doll slot routing can be isolated without brittle scene automation.

```bash
make client-unit
```

---

## Task 9 — Godot hotbar sync and equipment visuals

Files:
- Modify: `client/scripts/consumable_bar.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/equipment_visuals.gd`
- Modify: `client/scripts/smoke.gd`
- Modify: `client/tests/test_golden.gd`

- [x] Step 9.1: Replace local-only hotbar assignment state with snapshot/delta-driven `hotbar` and `hotbar_capacity`.
- [x] Step 9.2: Dragging a bag consumable onto a hotbar slot sends `assign_hotbar_intent`; clearing a slot sends null assignment.
- [x] Step 9.3: Hotkeys `1`-`9`, `0` call `use_hotbar_intent { slot_index }` only when the slot is enabled and assigned; disabled slots no-op client-side.
- [x] Step 9.4: Render disabled slots (`slot_index >= hotbar_capacity`) with reduced opacity/desaturation while preserving item assignment display.
- [x] Step 9.5: Preserve XP bar positioning below the hotbar without overlap.
- [x] Step 9.6: Migrate equipment visual resolver from `equipped.weapon` to `equipped.main_hand`.
- [x] Step 9.7: Add optional offhand shield/bow visual mapping only if it fits the current manifest pipeline without new production art; otherwise keep icons/ground loot presentation as the v28 minimum.
- [x] Step 9.8: Update smoke and golden tests for full equipment shape and hotbar fixture checks.

```bash
make client-unit
make client-smoke
```

---

## Task 10 — Godot client bot full equipment proof

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/tests/test_client_bot.gd`
- Create: `tools/bot/scenarios/client/10_full_equipment.json`
- Modify: existing `tools/bot/scenarios/client/*.json` as needed

- [x] Step 10.1: Add client bot actions for dragging bag items to named paper-doll slots and dragging equipped slots back to bag.
- [x] Step 10.2: Add assertions for paper-doll slot occupancy, tooltip stat text/debug data, hotbar capacity, hotbar assignment, and disabled slot styling.
- [x] Step 10.3: Add `assign_hotbar_slot` path that sends protocol-backed assignment rather than local-only `ConsumableBar.assign_slot`.
- [x] Step 10.4: Add `use_hotbar_slot` path using the new hotbar intent.
- [x] Step 10.5: Create `10_full_equipment.json`: open inventory, equip multiple slot types, assign hotbar, assert disabled slot presentation at capacity 2, equip belt to expand, use an enabled slot, and verify no local-only mutation before authoritative updates.
- [x] Step 10.6: Update existing client scenarios for `main_hand` migration and server-backed hotbar behavior.
- [x] Step 10.7: Add parser/validation tests for new client bot step types.

```bash
make client-unit
make bot-client
```

Verification run:

```bash
make client-unit
HEADLESS=1 make bot-client scenario=10_full_equipment
```

---

## Task 11 — Lifecycle docs and CI

Files:
- Modify: `docs/specs/v28_spec-full-equipment-and-belt-hotbar.md`
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v28_2026-06-07-full-equipment-and-belt-hotbar.md`

- [x] Step 11.2: Add v28 to the `PROGRESS.md` slice numbering note and lifecycle table.
- [x] Step 11.3: Add a concise "What v28 proved" section covering paper-doll equipment, two-hand rules, belt-gated hotbar persistence, droppable equipment templates, bot proofs, and client UI sync.
- [x] Step 11.4: Record deferred gaps: armor mitigation, block execution, affix grammar, comparison UI, stash/vendors, production icons/art, offhand abilities/dual-wield, and deeper dungeon drop economy.
- [x] Step 11.5: Mark this plan complete only after final CI is green.

```bash
make ci
```

---

## Final verification

- [x] `make validate-shared`
- [x] `make validate-assets`
- [x] `cd server && go test ./internal/store -run 'Hotbar|Equipment|CharacterItem|SessionStart'`
- [x] `cd server && go test ./internal/game/... -run 'Equip|Equipment|Hand|Occup|Ring|Bow|Weapon|Hotbar|Belt|UseHotbar|FullEquipment'`
- [x] `cd server && go test ./internal/http/... ./internal/replay/... -run 'Hotbar|Equipment|Replay|Reconnect|State|Session'`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make bot`
- [x] `make bot-client`
- [x] `make ci`

---

## Acceptance mapping

| Spec AC | Plan coverage |
|---------|---------------|
| 1 `make validate-shared` validates expanded contracts | Tasks 1-2, Final verification |
| 2 `weapon` gone from current schemas/examples | Task 1, Task 7, Task 9 |
| 3 Two-hand, bow, shield occupancy | Task 4, Task 7 |
| 4 Belt capacity 2..10 | Task 5, Task 9 |
| 5 Hotbar character persistence | Task 3, Task 6, Task 7 |
| 6 Disabled slots retain assignment and ignore hotkey | Task 5, Task 9, Task 10 |
| 7 Shield rolls display-only | Task 1, Task 8, Task 9 |
| 8 One template per category droppable | Task 1, Task 2, Task 7 |
| 9 Godot paper-doll through intents | Task 8, Task 10 |
| 10 Protocol/client bot proof | Task 7, Task 10 |
| 11 Legacy scenarios migrated; CI green | Task 7, Task 10, Task 11 |

---

## Deferred (explicit)

- Armor mitigation and block chance execution
- Affix grammar, procedural name generation, unique/set items
- Stash, vendors, gold wallet, crafting, loot filters, item comparison UI
- Production paper-doll art, production item icons, belt/hotbar VFX
- Offhand active abilities and dual-wield damage rules
- Expanding dungeon-wide treasure classes beyond the deterministic v28 equipment lab

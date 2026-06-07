# v23 Plan — Item templates and rolled drops

Status: Complete — `make ci` green on 2026-06-07
Goal: Add server-authoritative item templates, deterministic rolled weapon drops, rolled weapon damage, and rolled item tooltip presentation.
Architecture: Static item definitions remain for legacy consumables, currency, quests, and existing scenarios while new template entries roll concrete item instances through the Go Sim's seeded RNG. Rolled instance metadata travels as additive protocol fields and is persisted in v22's `rolled_stats` JSON so replay, reconnect, `/state`, and fresh sessions reconstruct the same item. The client remains presentation-only: it displays rarity/name/stats from the item instance and still sends only existing inventory intents.
Tech stack: Shared JSON rules/schemas/goldens, Go authoritative sim/realtime/http/replay/store, Python protocol bot, Godot inventory UI and data-only golden checks.

## Baseline and shortcut decision

v23 builds on v22 `character-scoped-persistence`: character item instances already persist across fresh sessions, session-start snapshots freeze replay inputs, and the store already round-trips `rolled_stats`. It reuses v21 generated dungeon mobs, v15 item presentation metadata, v13 inventory panel/tooltips, v8 equipped weapon damage, and existing protocol v1 snapshot/delta schemas.

Godot shortcut adoption checklist:

- **Decision:** reject plugin adoption.
- **Reason:** this slice only extends the existing inventory tooltip and placeholder item presentation for one template weapon. No new inventory grid, drag/drop model, camera system, production art, or addon-level UI is needed.
- **Borrow:** existing `inventory_panel.gd` tooltip/presentation pattern and `client/tests/test_golden.gd` data-only fixture checks. Existing placeholder sword visuals can represent `cave_blade`; no external asset pack is required.

Spec review notes resolved during planning:

- Slice number, branch, and baseline match `docs/PROGRESS.md`: v23 builds on completed v22.
- The spec's open questions have explicit defaults; this plan locks them for implementation: use `common`, `magic`, and `rare`; enforce only `required_level <= 1`; display/persist `max_hp` without applying it; expose enough rolled fields on loot entities for `/state` and future presentation; keep `items.v0.json`.
- Protocol changes are additive to existing v1 item/entity views and should not introduce a new envelope schema version unless validation proves one is required.
- Bot proof is mandatory because this touches gameplay, inventory, loot, protocol views, persistence, and replay; add scenario `16_rolled_drops.json`.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `shared/rules/item_templates.v0.schema.json` | Define template, rarity, requirements, rollable stats, and effect id contract |
| Add | `shared/rules/item_templates.v0.json` | Add first rolled weapon template(s), starting with `cave_blade` |
| Modify | `shared/rules/loot_tables.v0.schema.json` | Allow weighted entries keyed by exactly one of `item_def_id` or `item_template_id` |
| Modify | `shared/rules/loot_tables.v0.json` | Add `dungeon_mob_drop` template table while keeping legacy fixed drops |
| Modify | `shared/rules/monsters.v0.json` | Point `dungeon_mob` at rolled loot table |
| Modify | `shared/protocol/session_snapshot.v1.schema.json` | Add optional rolled item fields to inventory items and loot entities |
| Modify | `shared/protocol/state_delta.v1.schema.json` | Add optional rolled item fields to inventory item changes and loot entities |
| Modify | `shared/protocol/examples/session_snapshot.json` | Add representative rolled item example if validator examples need it |
| Modify | `shared/protocol/examples/state_delta.json` | Add representative rolled inventory/loot delta if validator examples need it |
| Add | `shared/golden/item_rolls.json` | Pin deterministic common and higher-rarity template roll cases |
| Modify | `tools/validate_shared.py` | Validate templates, template loot entries, protocol examples, and golden/rules drift |
| Modify | `server/internal/game/rules.go` | Parse templates, validate roll catalogs, and support template loot entries |
| Modify | `server/internal/game/types.go` | Extend `ItemView` and `EntityView` with optional rolled item metadata |
| Modify | `server/internal/game/sim.go` | Roll concrete item payloads, persist them in inventory/loot, and use rolled weapon damage |
| Modify | `server/internal/game/game_test.go` | Cover deterministic rolls, loot payloads, equip requirements, damage, and snapshots |
| Modify | `server/internal/realtime/runner.go` | Persist rolled payload JSON on character item add/update |
| Modify | `server/internal/realtime/hub.go` | Load session-start rolled payloads into fresh/replayed Sims if needed |
| Modify | `server/internal/replay/replay.go` | Preserve rolled item payloads in reconstruction and timeline snapshots |
| Modify | `server/internal/http/ws_test.go` | Assert rolled fields over WebSocket, reconnect, `/state`, and fresh sessions |
| Modify | `server/internal/http/replay_test.go` | Assert replay timeline includes rolled item metadata |
| Modify | `server/internal/store/models.go` | Add typed helper fields only if raw `RolledStats` is not enough |
| Modify | `server/internal/store/repos.go` | Preserve raw rolled payload; no schema change expected |
| Modify | `server/internal/store/store_test.go` | Extend existing `rolled_stats` round-trip with v23 payload shape |
| Modify | `client/scripts/inventory_panel.gd` | Render rarity, display name, and rolled stats from instance fields |
| Modify | `client/scripts/equipment_visuals.gd` | Treat template item ids as visual-compatible weapon defs if needed |
| Modify | `client/scripts/main.gd` | Preserve rolled loot/entity fields in client entity records if needed |
| Modify | `client/tests/test_golden.gd` | Add data-only item template/roll fixture checks |
| Modify | `tools/bot/run.py` | Add rolled item assertions and fresh-session checks |
| Modify | `tools/bot/test_protocol.py` | Unit-test new scenario assertions/helpers if added |
| Add | `tools/bot/scenarios/16_rolled_drops.json` | End-to-end rolled dungeon drop proof |
| Modify | `docs/PROGRESS.md` | Lifecycle update when v23 ships |

## Task 1 — Shared contracts and validation

Files:

- Add: `shared/rules/item_templates.v0.schema.json`
- Add: `shared/rules/item_templates.v0.json`
- Modify: `shared/rules/loot_tables.v0.schema.json`
- Modify: `shared/rules/loot_tables.v0.json`
- Modify: `shared/rules/monsters.v0.json`
- Modify: `shared/protocol/session_snapshot.v1.schema.json`
- Modify: `shared/protocol/state_delta.v1.schema.json`
- Modify: `shared/protocol/examples/session_snapshot.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Add: `shared/golden/item_rolls.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Define `item_templates.v0.schema.json` with `version`, `rarities`, and `templates`; require bounded `base_stats`, `rollable_stats`, `requirements`, and `effect_pool`.
- [x] Step 1.2: Add `cave_blade` as an equippable weapon template with `attack_mode`, `slot`, `reach`, base `damage_min`/`damage_max`, rollable damage stats, display-only `max_hp`, and empty `effect_pool`.
- [x] Step 1.3: Add rarity definitions for `common`, `magic`, and `rare`; keep weights and `stat_rolls` small enough for deterministic tests to inspect manually.
- [x] Step 1.4: Extend loot table schema entries to require `weight` and exactly one of `item_def_id` or `item_template_id`; keep `drops` and empty `no_drop` behavior valid for old scenarios.
- [x] Step 1.5: Add `dungeon_mob_drop` with a `cave_blade` template entry and change `dungeon_mob.loot_table` from `no_drop` to `dungeon_mob_drop`.
- [x] Step 1.6: Add optional protocol fields to inventory item views: `item_template_id`, `display_name`, `rarity`, `rolled_stats`, `requirements`, and `effect_ids`.
- [x] Step 1.7: Add optional rolled fields to loot `entity` views so `/state`, replay timeline, and future ground presentation can see the item before pickup.
- [x] Step 1.8: Add `shared/golden/item_rolls.json` with at least two pinned cases for `cave_blade`: one common roll and one magic or rare roll.
- [x] Step 1.9: Extend `tools/validate_shared.py` to validate template references, roll ranges, rarity weights, stat names, effect id arrays, requirements, template loot entries, and golden drift against rules.

```bash
make validate-shared
```

## Task 2 — Go rule loading and roll model

Files:

- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add typed `ItemTemplates`, `ItemTemplateDef`, `RarityDef`, `RollableStat`, `Requirements`, and `LootEntry.ItemTemplateID` models.
- [x] Step 2.2: Load `item_templates.v0.json` alongside existing shared rules and validate the same invariants as `tools/validate_shared.py`.
- [x] Step 2.3: Represent loot table resolution as a stable internal value that can return either a fixed item def or a template id; do not overload empty strings in observable behavior.
- [x] Step 2.4: Implement deterministic rarity selection and rollable stat selection using only `s.rng`, stable slice order from JSON arrays, and explicit tie-break rules.
- [x] Step 2.5: Roll stat values from inclusive integer ranges and merge them with base stats into the concrete item payload.
- [x] Step 2.6: Build display names with the rarity prefix plus template name; keep `item_def_id == item_template_id` for v23 compatibility.
- [x] Step 2.7: Add Go golden tests that load `item_rolls.json` and assert rarity, selected stats, stat values, display name, requirements, and effect ids.
- [x] Step 2.8: Add negative tests for unknown templates, invalid stat ranges, invalid rarity weights, malformed loot entries, and unsupported equip requirements.

```bash
cd server && go test ./internal/game/... -run 'ItemTemplate|ItemRoll|LootTable|Rules'
```

## Task 3 — Sim rolled loot and inventory payloads

Files:

- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 3.1: Extend internal loot and inventory item structs to carry a rolled item payload in addition to `itemDefID`.
- [x] Step 3.2: When `dropLoot` resolves a template entry, roll the concrete item once at monster death and attach that payload to the spawned loot entity.
- [x] Step 3.3: Ensure loot entity snapshots/deltas include `item_def_id`, `item_template_id`, `display_name`, `rarity`, rolled stats, requirements, and effect ids.
- [x] Step 3.4: On pickup, transfer the rolled payload from loot entity to inventory item; do not re-roll on pickup, equip, replay, or persistence load.
- [x] Step 3.5: Load persisted/session-start items with rolled payloads from `rolled_stats` JSON and advance `nextID` exactly as v22 already does.
- [x] Step 3.6: Preserve fixed item behavior for `rusty_sword`, `training_bow`, `training_badge`, `quest_leaf`, and `red_potion`.
- [x] Step 3.7: Add sim tests proving the same seed and ordered inputs produce identical loot entity payloads, inventory views, and final snapshots.
- [x] Step 3.8: Add tests proving a fixed-item loot table still drops a legacy item with no rolled fields.

```bash
cd server && go test ./internal/game/... -run 'Rolled.*Loot|Inventory.*Rolled|Snapshot.*Item|Legacy.*Loot'
```

## Task 4 — Rolled weapon equip and combat authority

Files:

- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 4.1: Update weapon lookup so a rolled equipped item with `damage_min` and `damage_max` uses that range before static `items.v0.json` damage.
- [x] Step 4.2: Keep weapon reach and attack mode resolved from the static item/template definition; rolled stats only affect v23 damage.
- [x] Step 4.3: Enforce only `required_level <= 1` for v23; reject unsupported or higher requirements with a clear intent rejection.
- [x] Step 4.4: Keep rolled `max_hp` display-only; add a test proving equipping a `max_hp` rolled item does not mutate player HP or max HP.
- [x] Step 4.5: Add combat tests proving rolled damage is authoritative for melee attacks and static damage remains the fallback for legacy weapons.
- [x] Step 4.6: Add ranged fallback coverage if `training_bow` paths share the same equipped weapon resolver.

```bash
cd server && go test ./internal/game/... -run 'Rolled.*Weapon|WeaponDamage|Equip.*Requirement|MaxHP'
```

## Task 5 — Persistence, replay, and HTTP protocol paths

Files:

- Modify: `server/internal/realtime/runner.go`
- Modify: `server/internal/realtime/hub.go`
- Modify: `server/internal/replay/replay.go`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/store/store_test.go`
- Modify: `server/internal/http/ws_test.go`
- Modify: `server/internal/http/replay_test.go`

- [x] Step 5.1: Serialize rolled item payloads into `store.CharacterItemInstance.RolledStats` on `inventory_add` and preserve them on equip/unequip updates.
- [x] Step 5.2: Treat empty or `{}` `rolled_stats` as a legacy fixed item with no rolled metadata.
- [x] Step 5.3: Load raw rolled payload JSON from character rows and session-start snapshots into the Sim without recomputing any roll.
- [x] Step 5.4: Add store tests for the full v23 rolled payload shape, not only the prior placeholder `{"prefix":"sharp"}` shape.
- [x] Step 5.5: Add WebSocket tests proving `inventory_add`, initial snapshot, reconnect snapshot, `/state`, and fresh same-account sessions include the same rolled item fields.
- [x] Step 5.6: Add replay tests proving timeline snapshots/deltas include rolled loot/inventory metadata and replay reconstruction preserves combat damage from the persisted rolled item.
- [x] Step 5.7: Confirm session-start snapshots isolate historical replay from later live character item mutations.

```bash
cd server && go test ./internal/store ./internal/http/... ./internal/replay/... -run 'Rolled|Item|Persistence|Replay|State|Snapshot'
```

## Task 6 — Bot scenario

Files:

- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`
- Add: `tools/bot/scenarios/16_rolled_drops.json`

- [x] Step 6.1: Add bot assertions for rolled inventory item fields: `item_template_id`, `rarity`, `display_name`, required rolled stat keys, and equipped weapon def/template.
- [x] Step 6.2: Add a runtime assertion that a `monster_damaged` event after equipping the rolled weapon meets the rolled damage lower bound when the scenario needs to prove damage use.
- [x] Step 6.3: Add fresh-session check support that asserts the same rolled item instance id and payload survive into a new same-account session.
- [x] Step 6.4: Create `16_rolled_drops.json`: start in `dungeon_levels`, descend, kill a `dungeon_mob`, pick up its rolled `cave_blade`, equip it, kill another mob or target with it, then assert rolled fields.
- [x] Step 6.5: Include `/state`, reconnect resume, replay, and fresh-session persistence checks through the existing bot runner flow.
- [x] Step 6.6: Migrate scenarios `12`, `13`, `14`, and `15` only as needed because `dungeon_mob` now drops loot; keep older scenarios from accidentally depending on `no_drop`.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
make db-up
make bot
```

## Task 7 — Godot client tooltip and golden checks

Files:

- Modify: `client/scripts/inventory_panel.gd`
- Modify: `client/scripts/equipment_visuals.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_golden.gd`

- [x] Step 7.1: Update `_tooltip` to accept the full item dictionary and prefer `display_name`, `rarity`, `requirements`, and `rolled_stats` from the item instance.
- [x] Step 7.2: Render rolled damage from `rolled_stats.damage_min` / `rolled_stats.damage_max` when present; fall back to static `items.v0.json` for legacy items.
- [x] Step 7.3: Display `max_hp` as a rolled stat without implying it changes current HP.
- [x] Step 7.4: Ensure `_is_weapon`, icons, ground loot, and equipped visuals work when `item_def_id` equals a template id; add minimal presentation mapping for `cave_blade` if existing sword visuals do not cover it.
- [x] Step 7.5: Preserve rolled fields in `main.gd` entity records if the client receives them on loot entities.
- [x] Step 7.6: Add data-only Godot golden checks that `item_rolls.json` references existing template rules and that tooltip-relevant fields are present; do not reimplement server roll simulation in GDScript.

```bash
make client-unit
make client-smoke
```

## Task 8 — Lifecycle docs and CI

Files:

- Modify: `docs/PROGRESS.md`

- [x] Step 8.1: When implementation ships, add v23 to the lifecycle table and mark latest completed slice as `item-templates-and-rolled-drops`.
- [x] Step 8.2: Document as-built behavior: template rules, rarity/stat rolls, dungeon mob drops, rolled weapon damage, rolled persistence, tooltip display, bot scenario `16`, and unchanged legacy fixed items.
- [x] Step 8.3: Record deferred follow-ups: affix grammar, item name generator, armor/jewelry/offhand, requirements beyond level 1, special-effect execution, comparison UI, stash/vendors/gold/crafting/trade, loot filters, and production art.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'ItemTemplate|ItemRoll|Rolled|Loot|WeaponDamage|Requirement'`
- [x] `cd server && go test ./internal/store ./internal/http/... ./internal/replay/... -run 'Rolled|Item|Persistence|Replay|State|Snapshot'`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make bot`
- [x] `make ci`

Manual check:

```bash
make play
# Descend into the dungeon, kill a dungeon_mob, pick up a rolled cave_blade,
# inspect the tooltip, equip it, restart play, and confirm the same rolled item persists.
```

## Deferred scope

- No full affix economy, prefix/suffix grammar, procedural item-name generator, or special-effect execution.
- No armor, rings, amulets, offhand, stash, crafting, vendors, gold, trade, item comparison UI, or loot filters.
- No character level/stat system beyond validating and rejecting unsupported requirements.
- No production item art or inventory plugin adoption.
- No Protobuf migration or non-additive protocol redesign.

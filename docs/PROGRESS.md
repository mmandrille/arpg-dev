# Project progress & slice lifecycle

**Read this file at the start of every new task** before writing specs, plans, or code.
It is the canonical snapshot of what exists, what each slice proved, and what is still open.

Last updated: 2026-06-08

---

## Current status

| Field | Value |
|-------|-------|
| **Latest completed slice** | v36 — `inventory-paper-doll-capacity` |
| **Active branch** | `main` |
| **CI gate** | `make ci` green on 2026-06-08 |
| **Next slice** | TBD |

### Slice numbering note

ADR-0001 sometimes calls the first slice **v1**; repo lifecycle labels use **v0–v9**
(`v0` = first playable). **Spec and plan filenames** use a `vN_` prefix for execution order:

```text
v1_* = first-playable    v5_* = resume-state    v8_* = equipped-weapon-damage
v2_* = equip-and-see-it  v6_* = visual-bot
v3_* = animate-and-react v7_* = gear-before-combat v9_* = solid-collision
v4_* = take-a-hit        v10_* = click-action-and-melee-range
v11_* = click-to-move-and-auto-path
v12_* = ranged-projectile-combat
v13_* = inventory-ui
v14_* = godot-client-bot
v15_* = item-visuals-and-loot-presentation
v16_* = use-consumable
v17_* = monster-chase-movement
v18_* = dungeon-levels-and-stairs
v19_* = teleporters-and-waypoint-ui
v20_* = play-session-loop
v21_* = dungeon-monster-combat
v22_* = character-scoped-persistence
v23_* = item-templates-and-rolled-drops
v24_* = main-menu-and-character-start
v25_* = treasure-classes-and-guarded-chests
v26_* = character-stats-and-leveling
v27_* = hold-click-controls
v28_* = full-equipment-and-belt-hotbar
v29_* = dungeon-equipment-drop-expansion
v30_* = monster-rarity-and-loot-scaling
v31_* = combat-stat-effects-and-feedback
v32_* = test-floor-and-resilient-scenarios
v33_* = true-coop-session
v34_* = model-reaction-polish
v35_* = boss-floor-gate
v36_* = inventory-paper-doll-capacity
```

Pattern: `docs/specs/vN_spec-<codename>.md`, `docs/plans/vN_<YYYY-MM-DD>-<codename>.md`.

---

## Slice lifecycle

Slices are small, end-to-end proofs. Each ships: shared contracts → Go sim → Godot client →
Python bot/smoke → golden fixtures → `make ci` green.

```text
v0 first-playable ──► v2 equip-and-see-it ──► v3 animate-and-react ──► v4 take-a-hit ──► v5 resume-state ──► v6 visual-bot-scenarios ──► v7 gear-before-combat ──► v8 equipped-weapon-damage ──► v9 solid-collision ──► v10 click-action ──► v11 auto-path ──► v12 ranged-projectiles ──► v13 inventory-ui
   (architecture)        (visual pipeline)         (skeletal anims)         (player damage)      (resume replay)      (visual replay playlist)        (world presets)              (weapon damage)             (walls + bodies)
        │                      │                        │                        │                         │                         │                              │                              │                         │
     main ✓                  main ✓                    main ✓                    main ✓              branch ✓                  branch ✓                       branch ✓                       branch ✓                  branch ✓                  branch ✓
```

| Slice | Codename | Status | Spec | Plan |
|-------|----------|--------|------|------|
| **v0** | `first-playable-vertical-slice` | Complete (on `main`) | [`v1_spec-first-playable-vertical-slice.md`](specs/v1_spec-first-playable-vertical-slice.md) | [`v1_2026-06-05-first-playable-vertical-slice.md`](plans/v1_2026-06-05-first-playable-vertical-slice.md) |
| **v2** | `equip-and-see-it` | Complete (on `main`) | [`v2_spec-equip-and-see-it.md`](specs/v2_spec-equip-and-see-it.md) | [`v2_2026-06-05-equip-and-see-it.md`](plans/v2_2026-06-05-equip-and-see-it.md) |
| **v3** | `animate-and-react` | Complete (on `main`) | [`v3_spec-animate-and-react.md`](specs/v3_spec-animate-and-react.md) | [`v3_2026-06-05-animate-and-react.md`](plans/v3_2026-06-05-animate-and-react.md) |
| **v4** | `take-a-hit` | Complete (on `main`) | [`v4_spec-take-a-hit.md`](specs/v4_spec-take-a-hit.md) | [`v4_2026-06-05-take-a-hit.md`](plans/v4_2026-06-05-take-a-hit.md) |
| **v5** | `resume-authoritative-state` | Complete (`make ci` green) | [`v5_spec-resume-authoritative-state.md`](specs/v5_spec-resume-authoritative-state.md) | [`v5_2026-06-05-resume-authoritative-state.md`](plans/v5_2026-06-05-resume-authoritative-state.md) |
| **v6** | `visual-bot-scenario-runner` | Complete (`make ci` green) | [`v6_spec-visual-bot-scenario-runner.md`](specs/v6_spec-visual-bot-scenario-runner.md) | [`v6_2026-06-05-visual-bot-scenario-runner.md`](plans/v6_2026-06-05-visual-bot-scenario-runner.md) |
| **v7** | `gear-before-combat-scenario` | Complete (`make ci` green) | [`v7_spec-gear-before-combat-scenario.md`](specs/v7_spec-gear-before-combat-scenario.md) | [`v7_2026-06-05-gear-before-combat-scenario.md`](plans/v7_2026-06-05-gear-before-combat-scenario.md) |
| **v8** | `equipped-weapon-damage` | Complete (`make ci` green) | [`v8_spec-equipped-weapon-damage.md`](specs/v8_spec-equipped-weapon-damage.md) | [`v8_2026-06-05-equipped-weapon-damage.md`](plans/v8_2026-06-05-equipped-weapon-damage.md) |
| **v9** | `solid-collision-and-obstacles` | Complete (`make ci` green) | [`v9_spec-solid-collision-and-obstacles.md`](specs/v9_spec-solid-collision-and-obstacles.md) | [`v9_2026-06-05-solid-collision-and-obstacles.md`](plans/v9_2026-06-05-solid-collision-and-obstacles.md) |
| **v10** | `click-action-and-melee-range` | Complete (`make ci` green) | [`v10_spec-click-action-and-melee-range.md`](specs/v10_spec-click-action-and-melee-range.md) | [`v10_2026-06-05-click-action-and-melee-range.md`](plans/v10_2026-06-05-click-action-and-melee-range.md) |
| **v11** | `click-to-move-and-auto-path` | Complete (`make ci` green) | [`v11_spec-click-to-move-and-auto-path.md`](specs/v11_spec-click-to-move-and-auto-path.md) | [`v11_2026-06-05-click-to-move-and-auto-path.md`](plans/v11_2026-06-05-click-to-move-and-auto-path.md) |
| **v12** | `ranged-projectile-combat` | Complete (`make ci` green) | [`v12_spec-ranged-projectile-combat.md`](specs/v12_spec-ranged-projectile-combat.md) | [`v12_2026-06-05-ranged-projectile-combat.md`](plans/v12_2026-06-05-ranged-projectile-combat.md) |
| **v13** | `inventory-ui` | Complete (`make ci` green) | [`v13_spec-inventory-ui.md`](specs/v13_spec-inventory-ui.md) | [`v13_2026-06-05-inventory-ui.md`](plans/v13_2026-06-05-inventory-ui.md) |
| **v14** | `godot-client-bot` | Complete (`make ci` green) | [`v14_spec-godot-client-bot.md`](specs/v14_spec-godot-client-bot.md) | [`v14_2026-06-02-godot-client-bot.md`](plans/v14_2026-06-02-godot-client-bot.md) |
| **v15** | `item-visuals-and-loot-presentation` | Complete (`make ci` green) | [`v15_spec-item-visuals-and-loot-presentation.md`](specs/v15_spec-item-visuals-and-loot-presentation.md) | [`v15_2026-06-06-item-visuals-and-loot-presentation.md`](plans/v15_2026-06-06-item-visuals-and-loot-presentation.md) |
| **v16** | `use-consumable` | Complete (`make ci` green) | [`v16_spec-use-consumable.md`](specs/v16_spec-use-consumable.md) | [`v16_2026-06-06-use-consumable.md`](plans/v16_2026-06-06-use-consumable.md) |
| **v17** | `monster-chase-movement` | Complete (`make ci` green) | [`v17_spec-monster-chase-movement.md`](specs/v17_spec-monster-chase-movement.md) | [`v17_2026-06-06-monster-chase-movement.md`](plans/v17_2026-06-06-monster-chase-movement.md) |
| **v18** | `dungeon-levels-and-stairs` | Complete (`make ci` green) | [`v18_spec-dungeon-levels-and-stairs.md`](specs/v18_spec-dungeon-levels-and-stairs.md) | [`v18_2026-06-06-dungeon-levels-and-stairs.md`](plans/v18_2026-06-06-dungeon-levels-and-stairs.md) |
| **v19** | `teleporters-and-waypoint-ui` | Complete (`make ci` green) | [`v19_spec-teleporters-and-waypoint-ui.md`](specs/v19_spec-teleporters-and-waypoint-ui.md) | [`v19_2026-06-06-teleporters-and-waypoint-ui.md`](plans/v19_2026-06-06-teleporters-and-waypoint-ui.md) |
| **v20** | `play-session-loop` | Complete (`make ci` green) | [`v20_spec-play-session-loop.md`](specs/v20_spec-play-session-loop.md) | [`v20_2026-06-06-play-session-loop.md`](plans/v20_2026-06-06-play-session-loop.md) |
| **v21** | `dungeon-monster-combat` | Complete (`make ci` green) | [`v21_spec-dungeon-monster-combat.md`](specs/v21_spec-dungeon-monster-combat.md) | [`v21_2026-06-06-dungeon-monster-combat.md`](plans/v21_2026-06-06-dungeon-monster-combat.md) |
| **v22** | `character-scoped-persistence` | Complete (`make ci` green) | [`v22_spec-character-scoped-persistence.md`](specs/v22_spec-character-scoped-persistence.md) | [`v22_2026-06-07-character-scoped-persistence.md`](plans/v22_2026-06-07-character-scoped-persistence.md) |
| **v23** | `item-templates-and-rolled-drops` | Complete (`make ci` green) | [`v23_spec-item-templates-and-rolled-drops.md`](specs/v23_spec-item-templates-and-rolled-drops.md) | [`v23_2026-06-07-item-templates-and-rolled-drops.md`](plans/v23_2026-06-07-item-templates-and-rolled-drops.md) |
| **v24** | `main-menu-and-character-start` | Complete (`make ci` green) | [`v24_spec-main-menu-and-character-start.md`](specs/v24_spec-main-menu-and-character-start.md) | [`v24_2026-06-07-main-menu-and-character-start.md`](plans/v24_2026-06-07-main-menu-and-character-start.md) |
| **v25** | `treasure-classes-and-guarded-chests` | Complete (`make ci` green) | [`v25_spec-treasure-classes-and-guarded-chests.md`](specs/v25_spec-treasure-classes-and-guarded-chests.md) | [`v25_2026-06-07-treasure-classes-and-guarded-chests.md`](plans/v25_2026-06-07-treasure-classes-and-guarded-chests.md) |
| **v26** | `character-stats-and-leveling` | Complete (`make ci` green) | [`v26_spec-character-stats-and-leveling.md`](specs/v26_spec-character-stats-and-leveling.md) | [`v26_2026-06-07-character-stats-and-leveling.md`](plans/v26_2026-06-07-character-stats-and-leveling.md) |
| **v27** | `hold-click-controls` | Complete (`make ci` green) | [`v27_spec-hold-click-controls.md`](specs/v27_spec-hold-click-controls.md) | [`v27_2026-06-07-hold-click-controls.md`](plans/v27_2026-06-07-hold-click-controls.md) |
| **v28** | `full-equipment-and-belt-hotbar` | Complete (`make ci` green) | [`v28_spec-full-equipment-and-belt-hotbar.md`](specs/v28_spec-full-equipment-and-belt-hotbar.md) | [`v28_2026-06-07-full-equipment-and-belt-hotbar.md`](plans/v28_2026-06-07-full-equipment-and-belt-hotbar.md) |
| **v29** | `dungeon-equipment-drop-expansion` | Complete (`make ci` green) | [`v29_spec-dungeon-equipment-drop-expansion.md`](specs/v29_spec-dungeon-equipment-drop-expansion.md) | [`v29_2026-06-07-dungeon-equipment-drop-expansion.md`](plans/v29_2026-06-07-dungeon-equipment-drop-expansion.md) |
| **v30** | `monster-rarity-and-loot-scaling` | Complete (`make ci` green) | [`v30_spec-monster-rarity-and-loot-scaling.md`](specs/v30_spec-monster-rarity-and-loot-scaling.md) | [`v30_2026-06-07-monster-rarity-and-loot-scaling.md`](plans/v30_2026-06-07-monster-rarity-and-loot-scaling.md) |
| **v31** | `combat-stat-effects-and-feedback` | Complete (`make ci` green) | [`v31_spec-combat-stat-effects-and-feedback.md`](specs/v31_spec-combat-stat-effects-and-feedback.md) | [`v31_2026-06-07-combat-stat-effects-and-feedback.md`](plans/v31_2026-06-07-combat-stat-effects-and-feedback.md) |
| **v32** | `test-floor-and-resilient-scenarios` | Complete (`make ci` green) | [`v32_spec-test-floor-and-resilient-scenarios.md`](specs/v32_spec-test-floor-and-resilient-scenarios.md) | [`v32_2026-06-08-test-floor-and-resilient-scenarios.md`](plans/v32_2026-06-08-test-floor-and-resilient-scenarios.md) |
| **v33** | `true-coop-session` | Complete (`make ci` green) | [`v33_spec-true-coop-session.md`](specs/v33_spec-true-coop-session.md) | [`v33_2026-06-08-true-coop-session.md`](plans/v33_2026-06-08-true-coop-session.md) |
| **v34** | `model-reaction-polish` | Complete (`make ci` green) | [`v34_spec-model-reaction-polish.md`](specs/v34_spec-model-reaction-polish.md) | [`v34_2026-06-08-model-reaction-polish.md`](plans/v34_2026-06-08-model-reaction-polish.md) |
| **v35** | `boss-floor-gate` | Complete (`make ci` green) | [`v35_spec-boss-floor-gate.md`](specs/v35_spec-boss-floor-gate.md) | [`v35_2026-06-08-boss-floor-gate.md`](plans/v35_2026-06-08-boss-floor-gate.md) |
| **v36** | `inventory-paper-doll-capacity` | Complete (`make ci` green) | [`v36_spec-inventory-paper-doll-capacity.md`](specs/v36_spec-inventory-paper-doll-capacity.md) | [`v36_2026-06-08-inventory-paper-doll-capacity.md`](plans/v36_2026-06-08-inventory-paper-doll-capacity.md) |

---

## What each slice proved

### v0 — First playable vertical slice

**Proves:** ADR-0001 architecture end-to-end.

- Go authoritative server + Godot thin client over JSON WebSocket
- Dev auth, solo session create/resume, Postgres persistence
- Deterministic 20 Hz sim (move, attack, loot drop, pickup, equip)
- Seeded replay + Python protocol bot + headless Godot smoke
- `GET /v0/sessions/{id}/state` inspection for agents

**Key as-built decisions:** session-scoped inventory (not character-scoped); WebSocket
`?access_token=` fallback; monster corpse at `hp == 0`; combat always hits in v0 (no range gate).

### v2 — Equip and see it

**Proves:** ADR-0001 D7 Tier A + ADR-0006 asset pipeline contract.

- Shared `item_visuals.v0.json` + `assets.v0.json` → Godot mount on `right_hand_socket`
- Deterministic `gen_glb.py` runtime assets; `make validate-assets`
- Equipped `rusty_sword` visible on character; server authority unchanged
- Resume restores equipped weapon from persisted inventory

**Scope limit:** only `rusty_sword` has a visual mapping; other items deferred.

### v3 — Animate and react

**Proves:** ADR-0007 animation state model; rigged GLB → skeletal clips pipeline.

- Player: `idle` / `walk` / `attack` from client input/prediction; weapon on `hand_r` bone
- Monster: `hit` / `death` from authoritative `monster_damaged` / `monster_killed` events
- `AnimationController` priority machine: terminal > one-shot > locomotion
- Clips built by `client/tools/build_animations.gd` → committed `.tres` libraries
- **No server/protocol change** — client starts reading existing `state_delta.events`

### v4 — Take a hit

**Proves:** Bidirectional combat + player reactions on the same event-driven path as monsters.

- Per-monster optional `retaliation_damage` in `shared/rules/monsters.v0.json`
- Server emits `player_damaged` / `player_killed`; dead player intents rejected
- Client: player `hit` / `death` clips; input gated when `hp <= 0`
- Golden: `pinned_seed` `deadbeefdeadbeef` → `final_player_hp: 9` (one hit, one retaliation)
- Bot/smoke assert `hp < 10` on random seeds (not exact golden HP)
- Extras: `test_golden.gd` retaliation gate; `make bot-visual` for interactive inspection

**Explicit non-goals (still true):** no respawn, no healing, no monster attack anim on retaliate.

### v5 — Resume authoritative state

**Proves:** Same-session reconnect restores server-owned combat/world state through deterministic replay.

- WebSocket resume uses `replay.Reconstruct` when `session_inputs` exist.
- No `LoadInventory` on replay resume; inventory/equipped state comes from recorded pickup/equip inputs.
- Initial resume snapshot restores player HP, monster HP/death, inventory, equipped weapon, server tick, and ID continuity.
- Runner seeds historical message IDs and next sequence, so old intents reject as `duplicate`.
- Bot and Godot smoke assert real resumed monster death and reduced player HP.
- Extras: dead-player resume rejects gameplay intents; `/state` and WebSocket resume snapshot parity.

**Explicit non-goals (still true):** no character-scoped inventory, no respawn/healing/checkpoints, no protocol schema bump.

### v6 — Visual bot scenario runner

**Proves:** Bot scenarios are discoverable local artifacts and can be visually replayed without hardcoding Godot-only flows.

- `tools/bot/scenarios/*.json` defines declarative scenario steps and named assertions.
- `make bot` runs every discovered scenario through auth + WebSocket, then verifies `/state`, reconnect resume, and replay.
- `tools.bot.run --write-manifest` writes `.artifacts/bot-runs/*.json` with scenario/session metadata.
- Debug endpoint `GET /v0/sessions/{id}/replay/timeline` emits protocol-shaped snapshot/delta envelopes from deterministic replay.
- `make bot-visual` records all scenarios, verifies replay, then launches Godot with a visual replay playlist.
- Godot visual replay mode consumes the manifest and timeline envelopes through existing snapshot/delta render handlers.
- The visual replay client exits normally after the playlist completes; set `ARPG_VISUAL_REPLAY_EXIT_ON_COMPLETE=0` to keep it open.

**Explicit non-goals (still true):** no production replay browser, no durable artifact retention policy, no client presentation annotations beyond authoritative events.

### v7 — Gear before combat scenario

**Proves:** The server can own multiple deterministic initial world presets, and replay/resume/debug timelines reconstruct the selected preset instead of drifting to the default world.

- Shared `worlds.v0.json` defines `vertical_slice` and `gear_before_combat` initial layouts.
- Sessions persist `world_id`; create defaults to `vertical_slice`, rejects unknown worlds, and resume returns the persisted world.
- `game.NewSimWithWorld` spawns the player, initial loot, and monsters from rules data; `NewSim` remains a default wrapper.
- Replay reconstruction, `/state`, replay timeline, and WebSocket fresh/resume paths use the persisted world.
- Bot scenario catalog now runs `01_vertical_slice.json` then `02_gear_before_combat.json`.
- Gear scenario walks to initial `rusty_sword`, picks it up, equips it, kills `training_dummy_reward`, picks up `training_badge`, and asserts two inventory items.

**Explicit non-goals (still true):** no pickup range gate, no `world_id` in WebSocket snapshots, no Godot inventory UI for non-visual items.

### v8 — Equipped weapon damage

**Proves:** Equipped item rules can change authoritative combat outcomes without protocol, replay, or client UI changes.

- `rusty_sword` declares `damage: {min: 3, max: 5}` in `shared/rules/items.v0.json`.
- Server attack damage resolves the equipped weapon at hit time; missing/no-damage equipment falls back to `combat.player_damage`.
- Go and GDScript golden tests consume `shared/golden/equipped_weapon_damage.json`.
- `tools/validate_shared.py` rejects damage on non-weapon or non-equippable items and checks golden/rules drift.
- `gear_before_combat` now asserts `training_dummy_reward` dies in one acknowledged equipped attack.
- Replay, reconnect resume, `/state`, and Godot smoke stay green through `make ci`.

**Explicit non-goals:** no additive stat system, armor, healing, client damage preview, or inventory
UI/plugin adoption. Attack range was deferred in v8 and closed by v10.

### v9 — Solid collision and obstacles

**Proves:** The authoritative server can block player movement against live monster bodies and
static world walls while preserving replay/resume determinism.

- Shared `worlds.v0.json` now supports static `wall` entries with axis-aligned rectangular sizes.
- `collision_lab` world places wall obstacles with a middle passage and a live monster beyond them.
- Server movement checks player circle vs live monster circles and wall AABBs; diagonal moves slide
  on one axis when possible.
- Dead monsters are non-solid, so corpses do not block loot/combat scenario flow.
- Python bot adds `move_until_player_position` and a collision lab scenario proving traversal
  through the wall gap before the final monster attack, `/state`, reconnect, and replay.
- Godot renders simple static wall boxes from shared world rules for fresh sessions and visual replay
  manifests; the server still owns all collision outcomes.
- `make ci` green on 2026-06-05.

**Explicit non-goals:** no pathfinding, navmesh, monster movement/AI, polygon collision, or wall
protocol entities. Attack range was deferred in v9 and closed by v10.

### v10 — Click action and melee range

**Proves:** A single left-click action can cover combat, loot pickup, and interactable activation
while the server enforces melee reach and mutable world object state deterministically.

- `action_intent { target_id }` replaces active `attack_intent` / `pick_up_intent` protocol use.
- Shared combat/item rules define `combat.unarmed_reach` and weapon `reach`; Go and GDScript
  consume `shared/golden/melee_reach.json`.
- Server rejects in-world actionable targets beyond reach with `out_of_range`.
- `wooden_door` interactables spawn from shared rules, block movement while closed, open through
  an authoritative action, emit `interactable_activated`, and unblock passage.
- Godot left-click ray-picks monsters, loot, and doors through per-entity pick colliders; doors are
  rendered as simple in-repo panels that tween open from authoritative state.
- Bot scenarios `01`-`03` now use action steps; `04_door_lab` proves far reject, door open,
  passage, loot pickup, reconnect resume, and replay.
- `make ci` green on 2026-06-05.

**Explicit non-goals (still true):** no click-to-move, pathfinding, ranged weapons, key/lock
puzzles, door closing, inventory UI, or production door art.

### v11 — Click to move and auto path

**Proves:** The server can own deterministic click-to-move and auto-approach using shared
navigation rules while preserving replay/resume behavior.

- `move_to_intent { position }` queues server-owned floor-click movement.
- Out-of-range `action_intent` plans to a reachable melee approach cell, queues movement, and
  executes the original action on arrival with one acceptance ack.
- Shared `navigation.v0.json` defines `cell_size`, `max_auto_steps`, search bounds, and
  `stop_distance`; `auto_path.json` pins the path-maze approach fixture.
- Go A* rasterizes walls, live monsters, and closed interactables from the same collision rules
  used by movement; manual `move_intent` cancels queued auto-navigation.
- `path_maze` world plus bot scenario `05_path_maze.json` proves one entity click routes through
  a wall maze and kills a target without scripted waypoints.
- Godot empty-floor left click sends `move_to_intent`; entity click stays `action_intent`.

**Explicit non-goals (still true):** no NavMesh authority, monster AI/pathfinding, path preview UI,
door closing, inventory UI, or production navigation polish.

### v12 — Ranged projectile combat

**Proves:** Ranged weapons can use server-owned traveling projectile entities with deterministic
impact-time collision, hit, damage, replay, and client presentation.

- `training_bow` declares `attack_mode: "ranged"`, weapon damage, reach, and projectile speed in
  shared item rules, with schema and validation guards.
- Ranged monster `action_intent` spawns a wire-visible `projectile` entity; melee combat, loot, and
  interactables keep their existing behavior.
- Projectile flight advances at 20 Hz and sweeps against inflated wall/door AABBs and live monster
  circles using nearest-hit selection with deterministic tie-breaks.
- Ranged hit chance and damage roll only at impact; miss emits `attack_missed` without retaliation.
- `ranged_projectile.json` pins gap kill, wall block, and miss/no-retaliation cases for Go and
  GDScript fixture checks.
- `ranged_lab` plus bot scenario `06_ranged_lab.json` proves bow pickup/equip, ranged kill beyond
  melee range through a wall gap, `/state`, reconnect resume, and replay.
- Godot renders placeholder projectile entities from authoritative spawn/update/remove deltas.
- `make ci` green on 2026-06-05.

**Explicit non-goals for v12:** no spells, piercing, homing/AoE, monster ranged AI, predictive
leading, ranged pickup/door activation, production bow art, inventory UI, or projectile catalog.

### v13 — Inventory UI

**Proves:** Human-facing inventory presentation can stay display-only while server-owned inventory
intents mutate authoritative state, persistence, replay, and resume.

- `unequip_intent` and `drop_intent` extend the protocol; `inventory_remove` lets deltas remove bag
  rows without waiting for a fresh snapshot.
- Server drop placement is deterministic, collision-free, adjacent to the player, and pinned by
  `shared/golden/inventory_drop.json` in Go and GDScript fixture checks.
- Dropping an equipped item clears `equipped.weapon`, removes the inventory row, spawns pickup-able
  loot, persists the removal, and reconstructs through replay/resume.
- `inventory_lab` plus bot scenario `07_inventory_lab.json` proves pickup, equip, unequip, drop,
  re-pickup, and re-equip over protocol, `/state`, reconnect resume, and replay.
- Godot adds a custom Diablo-dark panel toggled with `I`, one weapon slot, a bag grid, tooltips from
  item rules, double-click/drag equip, drag-to-bag unequip, drag-outside drop, and no local inventory
  authority.
- The old `Q` equip shortcut and debug hints are removed; autoplay and bot continue using explicit
  protocol `equip_intent`.
- `make ci` green on 2026-06-05.

**Explicit non-goals (still true):** no stash, vendors, crafting, stack splitting, equipment slots
beyond weapon, production item icons, Godot inventory plugins, character-scoped persistence, item
destruction, or drop targeting/range gates.

### v14 — Godot client bot

**Proves:** The client input pipeline (ray-pick targeting, inventory UI, keyboard shortcuts) can be
driven and asserted by an automated bot running inside `main.tscn` in headless Godot, in CI,
without a human watching.

- `BotController` mounts inside `main.tscn` when `ARPG_BOT_CLIENT=1`; `BotScenarioRunner`
  executes client scenarios one frame-tick step at a time.
- `get_bot_state()` exposes reconciled client state (ws_open, entities, inventory, equipped,
  pending_events) as a read-only dictionary; the bot dispatches intents through `bot_dispatch_action`
  and `bot_dispatch_inventory_intent` which route through the same `client.send()` and
  `_on_inventory_intent_requested()` paths as human input.
- `press_key KEY_I` pushes a real `InputEventKey` through `get_viewport().push_input()` and
  toggles the actual `InventoryPanel` via `_unhandled_input()`.
- Headless ray-pick fallback: `click_entity` dispatches `action_intent` directly (documented fallback;
  `Input.warp_mouse()` has no effect without a real display server, making `get_mouse_position()`
  unreliable for ray-pick targeting in `--headless` mode).
- `scripts/bot_client.sh` discovers `tools/bot/scenarios/client/*.json`, validates each, and runs
  one fresh Godot headless process per scenario, checking for the `[bot-client] PASS` sentinel.
- 5 client scenarios: `click_to_kill`, `inventory_open_close`, `inventory_equip_unequip`,
  `inventory_lab_drop_item`, `click_to_move` — all green against a live server.
- 24 `test_client_bot.gd` unit tests cover scenario parsing, validation, timeout messages, and
  PASS/FAIL sentinel formatting without requiring a live server — wired into `make client-unit`.
- `make bot-client` added to `make/agents.mk`; step 7/8 added to `scripts/ci.sh`; all 8 CI steps green.
- Python bot, replay verification, and visual replay are unchanged.

**As-built headless constraint:** `Input.warp_mouse()` is a no-op in `--headless` mode (no display
server); `click_entity` and `click_floor` therefore use the documented direct fallback (same
`action_intent`/`move_to_intent` WebSocket send path). A manual windowed run confirms ray-pick works
correctly with a real display. Drag-and-drop inventory operations also use the direct path through
`bot_dispatch_inventory_intent`, since Control drag events require a real display server.

**Explicit non-goals:** no multi-scenario concurrency per process, no pixel-level assertion,
no competing/multiplayer bots, no headless ray-pick workaround, no v14 changes to Go server
or shared protocol.

### v15 — Item visuals and loot presentation

**Proves:** Current item presentation can be shared-data-driven without making the client
authoritative for inventory, loot, or equipment outcomes.

- `shared/assets/item_presentations.v0.json` defines client-only icon and ground-loot presentation
  metadata for every current item rule: `rusty_sword`, `training_bow`, `training_badge`,
  `quest_leaf`, and `red_potion`.
- `make validate-shared` schema-validates the presentation file and cross-checks that every item
  rule has presentation metadata and no stale presentation keys exist.
- Godot inventory slots draw distinct shape/color icons from shared presentation data instead of
  text initials, while tooltips still resolve names/stats from item rules.
- Godot ground loot renders distinct primitive silhouettes for sword, bow, badge/coin, leaf, and
  potion from the same presentation metadata; missing metadata falls back to category coloring.
- Equipped weapon GLB mounting remains unchanged through `item_visuals.v0.json` and
  `assets.v0.json`; the server/protocol are unchanged.
- Godot client bot now asserts loot and inventory presentation metadata on the inventory drop
  scenario; `test_item_visuals.gd` checks presentation coverage for every current item.
- `make ci` green on 2026-06-06.

**Explicit non-goals:** no production art, imported icon pack, texture budget, Blender export
pipeline, remote patcher, stash, vendors, crafting, consumable use, or new gameplay item stats.

### v16 — Use consumable

**Proves:** Consumable item use can mutate authoritative HP and inventory while the hotbar stays a
client-only input surface.

- `use_intent { item_instance_id }` is decoded, persisted with the input stream, and resolved by
  the Go sim against server-owned inventory and item rules.
- `red_potion` declares `heal: { min: 5, max: 5 }`; `shared/golden/use_consumable.json` pins heal
  amount and HP cap behavior for Go/GDScript drift checks.
- Server removes the consumed inventory row, emits `item_used` and `player_healed`, and updates the
  player entity HP; rejects include non-consumable, missing item, full HP, and dead player cases.
- `heal_lab` plus protocol bot scenario `08_heal_lab.json` proves pickup of two potions, damage,
  two uses, `/state`, reconnect resume, and replay.
- Godot adds a bottom-center `ConsumableBar` with 10 client-only slots; drag assignment and keys
  `1`-`9`/`0` send `use_intent` for the assigned inventory item.
- Client bot scenario `06_use_potion_hotbar.json` exercises hotbar assignment, double-click bag
  use, key use, and inventory removal.
- `make ci` green on 2026-06-06.

**Explicit non-goals:** no server-side hotbar persistence, stack splitting, cooldowns, buffs,
heal-over-time, production potion art, stash, vendors, or crafting.

### v17 — Monster chase movement

**Proves:** opt-in server-authoritative monster chase with aggro, leash return, and v11 path reuse.

- Shared `behavior: "chase"` on `training_dummy_chase` with `aggro_radius`, `leash_radius`, and
  `move_speed == navigation.cell_size`; all legacy monsters default to static.
- `Sim.Tick` runs `advanceMonsterMovement` after player movement and before projectiles; monsters
  replan each tick, path around walls/player/other monsters, and emit edge-only `monster_aggro` /
  `monster_leashed` events plus `entity_update` position deltas.
- Golden `shared/golden/monster_chase.json` pins maze chase and leash return on seed
  `cafebabecafebabe`.
- Worlds `chase_lab`, `chase_maze`, and `leash_lab` plus bot scenarios `09`–`11` prove open-field
  chase, maze routing, and leash reset through `/state`, reconnect, and replay.
- Godot drives monster `walk`/`idle` from authoritative position deltas; `monster_anims.tres` adds
  a minimal walk clip.
- `make ci` green on 2026-06-06.

**Explicit non-goals:** no proactive monster attacks, behavior trees/LimboAI, group aggro, fractional
chase speeds, or NavMesh authority.

### v18 — Dungeon levels and stairs

**Proves:** The authoritative Sim can hold multiple generated dungeon levels and move the player
between them with deterministic, level-scoped deltas.

- `dungeon_levels` world runs in multi-level mode; legacy worlds remain single-level at level `0`.
- `LevelState` owns per-level entities, walls, movement, auto-nav, and navigation bounds.
- `shared/rules/dungeon_generation.v0.json` drives 32x20 dungeon floors, perimeter walls, level
  names, player spawn, and deterministic stair placement.
- `descend_intent` / `ascend_intent` move the player between generated levels and emit old-level
  remove + new-level full spawn deltas with `level_changed`.
- `shared/golden/dungeon_stairs.json` pins level -1/-2 stair and loot positions.
- Godot renders generated dungeon walls, placeholder stairs, and a top-right level HUD.

**Explicit non-goals:** no character-scoped persistence, town, waypoints, full room/corridor PCG,
monster density by depth, co-op routing, or production stair art.

### v19 — Teleporters and waypoint UI

**Proves:** Dungeon levels can expose session-scoped discovered teleporters and use them for
server-authoritative fast travel.

- Dungeon generation places one deterministic `teleporter` interactable per generated level.
- `action_intent` on a reachable teleporter discovers that level and emits
  `teleporter_discovery_update` plus `teleporter_discovered`.
- v1 snapshots include `discovered_teleporters`, listing generated/visited levels as enabled or
  disabled.
- `teleport_intent { target_level }` validates current teleporter reach and target discovery, then
  reuses v18 two-delta level transition output.
- Godot renders a placeholder teleporter and opens a left-side waypoint panel with disabled
  undiscovered rows and a scroll container for longer level lists.
- Bot scenario `13_teleporter_lab.json` covers discover -1, descend, verify -2 disabled, discover
  -2, and teleport back to -1.

**Explicit non-goals:** no character-scoped waypoint persistence, town waypoint, VFX/audio,
production teleporter art, hidden infinite level catalog, or plugin adoption.

### v20 — Play session loop

**Proves:** The generated dungeon can be entered from a static town and used as the default fresh
interactive play loop without changing the authoritative client/server boundary.

- `dungeon_levels` now starts at town level `0`, built from `worlds.v0.json` with a down stair and
  a teleporter; level `-1` is generated lazily on first descent.
- Town teleporter discovery is initialized server-side and appears in snapshots as level `0`
  discovered; protocol v1 now allows `target_level: 0`.
- Generated level `-1` now has a `stairs_up` landing at the dungeon player spawn, so
  `0 -> -1 -> -2 -> -1 -> 0` is replayable and golden-tested.
- `ascend_intent` from level `-1` returns to town at the town down stair; teleporting to town lands
  at the town teleporter when the current floor has an active discovered teleporter.
- `scripts/play.sh` launches a fresh `dungeon_levels` run by default, and the interactive Godot
  client requests `dungeon_levels` when no world is specified.
- Godot renders dungeon perimeter walls only below level `0`; town remains an open placeholder hub
  with the existing waypoint panel and level-HUD behavior.
- Bot scenarios `12_dungeon_levels` and `13_teleporter_lab`, replay goldens, and client golden
  checks were updated for the town preamble and town waypoint row.

**Explicit non-goals:** no character-scoped persistence, player-facing resume, safe-zone combat
rules, NPCs/vendors/stash, production town art, or plugin adoption.

### v21 — Dungeon monster combat

**Proves:** Generated dungeon floors can be dangerous without changing the authoritative
client/server boundary or adding client-side combat authority.

- Shared monster rules define `dungeon_mob` as a chase monster with proactive melee attack damage
  and tick-based cooldown.
- Dungeon generation places deterministic `dungeon_mob` entities on negative dungeon levels only;
  town level `0` remains monster-free.
- Server `Sim.Tick` advances monster chase, then proactive monster attacks, then projectiles,
  preserving deterministic replay order.
- Proactive attacks emit existing `player_damaged` / `player_killed` events, so current Godot
  player hit/death reactions work without a protocol schema change.
- `shared/golden/dungeon_monster_attack.json` pins seed, level, monster def, first damage tick,
  damage, and resulting HP for Go and Godot golden checks.
- Bot scenario `14_dungeon_monsters.json` proves descend, passive damage, dungeon mob kill,
  `/state`, reconnect resume, and replay.

**Explicit non-goals:** no monster loot drops beyond `no_drop`, no monster attack animation,
no depth scaling, no ranged/AoE monsters, no protocol-level town safe-zone guard, no
character-scoped persistence, and no production monster art.

### v22 — Character-scoped persistence

**Proves:** Default-character item instances, equipped weapon state, and waypoint unlocks can
survive fresh sessions while replay remains pinned to a session-start progression snapshot.

- Postgres now has character-owned item instances with `location`, `equipped`, `slot`, and
  future-ready `rolled_stats`, plus character waypoint rows keyed by level.
- Fresh session creation freezes the character's current items and waypoints into immutable
  session-start snapshot tables; WebSocket fresh attach, `/state`, replay, and timeline all load
  that snapshot before applying session inputs.
- Live inventory add/update/remove changes persist against the session character; dropped and
  consumed items are removed durably for v22.
- Teleporter discovery changes persist as character waypoints; town level `0` remains always
  available even when not explicitly stored.
- Same-session reconnect continues to reconstruct from recorded inputs, not mutable live
  character rows, so historical replay does not drift after later fresh-session progression.
- Bot scenario `15_character_persistence.json` proves gear/equipment persistence, persisted
  level `-1` waypoint access, fresh-session level generation, `/state`, reconnect, and replay.

**Explicit non-goals:** no character picker, player-facing old-session resume, stash UI,
vendors/gold/crafting/quests, character stats/skills/XP, respawn/checkpoints, durable dungeon
maps/monsters/floor drops/HP, or random item stat generation.

### v23 — Item templates and rolled drops

**Proves:** Dungeon kills can produce deterministic rolled gear that remains server-authoritative
through pickup, equip, combat, persistence, reconnect, fresh sessions, and replay.

- Shared `item_templates.v0.json` defines `cave_blade`, rarity weights, bounded rollable stats,
  requirements, and reserved effect ids as data.
- Loot tables now support entries keyed by exactly one of `item_def_id` or `item_template_id`;
  legacy fixed drops and empty `no_drop` remain valid.
- `dungeon_mob` now uses `dungeon_mob_drop`, rolling a concrete `cave_blade` payload at monster
  death with the seeded Go RNG.
- Rolled item metadata is additive in protocol v1 item and loot entity views: `item_template_id`,
  `display_name`, `rarity`, `rolled_stats`, `requirements`, and `effect_ids`.
- Character item persistence stores the durable rolled payload in v22's `rolled_stats` JSON and
  reloads it through session-start snapshots without re-rolling.
- Equipped rolled weapons use rolled `damage_min` / `damage_max` for authoritative damage; rolled
  `max_hp` is display-only in v23.
- Godot inventory tooltips display instance rarity, display name, rolled damage, `max_hp`, and
  requirements; `cave_blade` reuses placeholder blade visuals.
- Bot scenario `16_rolled_drops.json` proves dungeon mob kill, rolled drop pickup, equip, damage
  use, `/state`, reconnect, replay, and fresh-session persistence.

**Explicit non-goals:** no affix grammar, procedural name generator, armor/jewelry/offhand,
stash/crafting/vendors/gold/trade, special-effect execution, item comparison UI, loot filters,
production item art, character stat requirements beyond level `1`, or Protobuf migration.

### v24 — Main menu and character start

**Proves:** The Godot client can boot into a player-facing shell while the server remains
authoritative for accounts, characters, ownership checks, and fresh-session bootstrap.

- Authenticated HTTP APIs list and create account-scoped named characters; duplicate display names
  are allowed in v24 and names are trimmed/length-limited server-side.
- Fresh session creation accepts an optional selected `character_id`, rejects cross-account
  character use, and preserves the default-character path for bots, smoke, replay, and dev flows.
- Interactive Godot startup now opens a main menu with Continue, New Game, Settings, and Exit;
  Continue starts a fresh `dungeon_levels` session from selected character progression.
- New Game creates a named character and starts a fresh `dungeon_levels` session for that
  character; old-world/session resume remains dev/debug-only.
- Local settings persist a fixed window size (`1280x720`, `1600x900`, `1920x1080`) in
  `user://settings.json` and apply immediately.
- ESC opens a pause menu with Resume, Settings, Return to Main Menu, and Exit; overlay visibility
  blocks gameplay clicks, WASD, hotbar, inventory, camera zoom, and bot-dispatched gameplay intents.
- Return to Main Menu closes the WebSocket, marks the session ended through a small idempotent
  owner-only route, clears client gameplay state, and offers only fresh character starts.
- Client bot scenario `08_main_menu_flow.json` proves settings, named character creation, pause
  input lock, return to menu, Continue, and fresh new session id; scenarios `01`-`07` keep their
  explicit auto-start path.

**Explicit non-goals:** no character delete/rename/class/customization/portraits, production menu
art/audio, richer settings, old-session resume UI, character summaries, stash/vendors/quests, or
durable dungeon maps/monsters/floor drops/HP/current level.

### v25 — Treasure classes and guarded chests

**Proves:** Monster and chest rewards can resolve through data-driven treasure classes with
multiple ordered drop attempts, while rare procedural chests create guarded dungeon floors.

- Shared `treasure_classes.v0.json` defines ordered attempts with success/no-drop weights and
  weighted fixed item or item-template entries.
- `dungeon_mob_drop` now bridges through `dungeon_mob_tc_1`; its primary attempt produces a rolled
  `cave_blade`, while a lower-probability secondary attempt can add `red_potion` or the money-like
  `training_badge`.
- `guarded_chest_drop` bridges through `guarded_chest_tc_1`, giving chests a primary reward and a
  lower-probability bonus attempt.
- Dungeon generation has rare `chest_placement`; successful chest floors spawn a `treasure_chest`
  and apply `monster_count_bonus` on that same level.
- Chest generation uses a labeled seed substream, so no-chest floors preserve existing stair and
  monster generation expectations.
- `treasure_chest` opens via existing `action_intent`, emits existing interactable/loot events,
  rolls loot once, and rejects repeated opens without duplicating drops.
- Bot scenario `17_treasure_classes_and_guarded_chests.json` pins a guarded chest floor, proves
  monster treasure-class loot, chest open-once behavior, pickup, `/state`, reconnect, replay, and
  fresh-session persistence.
- Bot create-session now supports an optional pinned seed only in local development; normal remote
  sessions keep server-generated OS-entropy seeds.
- Gold wallet, Magic Find, unique/set catalogs, depth-banded treasure classes, boss-floor chest
  integration, and production chest art remain deferred.

**Explicit non-goals:** no Magic Find stat or rarity modifier, no unique/set catalogs, no real gold
wallet, no boss-floor rules, no production chest art/animation/audio, and no client-side drop logic.

### v26 — Character stats and leveling

**Proves:** Character-owned XP, levels, stat points, and derived substats can be durable,
server-authoritative progression while the Godot client remains a renderer/input surface.

- Shared `character_progression.v0.json` defines base stats, a table XP curve, points per level,
  and bounded derived-stat formulas for damage, armor, attack speed, hit chance, crit chance,
  crit damage, movement speed, HP, and mana.
- Dungeon mobs now award positive XP; monster kill XP applies exactly once, crosses level
  thresholds in order, and grants 5 unspent stat points per level.
- `character_progression` persists per character, and session-start progression snapshots preserve
  deterministic reconnect/replay boundaries.
- `allocate_stat_intent` is server-authoritative; invalid stats, dead-player allocation, and
  overspending reject without mutating state.
- `vit` allocation updates derived `max_hp` and raises current HP by the gained max; the first
  `str` damage hook adds derived damage to melee/fixed weapon damage.
- Armor, crit, hit chance, attack speed, movement speed, max mana, and magic damage are computed
  and displayed but remain gameplay-deferred.
- Godot adds a left-side `C`-toggle character sheet with stat `+` buttons and a compact XP bar
  below the hotbar; client state only updates after authoritative snapshots/deltas.
- Protocol bot scenario `18_character_stats_and_leveling.json` proves XP, level-up points, VIT
  allocation, overspend rejection, `/state`, replay, reconnect, and fresh-session persistence.
- Client bot scenario `09_character_stats_panel.json` proves the stats panel, XP bar, pause/menu
  allocation lock, VIT spend through the `+` button, and max HP UI update.

**Explicit non-goals:** no passive skill tree, no respec, no class selection, no stat requirements,
no mana consumers, no armor/crit/hit/attack-speed gameplay, and no main-menu character summaries.

### v27 — Hold click controls

**Proves:** Diablo-style sustained left-click input can live entirely in the Godot client by repeating
existing intents at the current send cadence, without protocol or server changes.

- Hold LMB on a live monster locks a sticky target and repeats `action_intent` at `SEND_INTERVAL`
  until the monster dies, the player dies, LMB releases, or the target becomes invalid.
- Hold LMB on floor repeats `move_to_intent` toward the mouse ground point when cursor movement
  exceeds a 0.25 xz epsilon.
- Loot, doors, stairs, teleporters, and chest clicks stay one-shot; open chests are non-actionable
  and do not spam intents.
- Out-of-range hold-attack still uses v11 auto-approach; WASD cancel of auto-nav is unchanged.
- `SustainedClickInput` helper + `test_sustained_input.gd` cover hold start/stop/epsilon logic;
  bot hold+drag scenario remains deferred.

**Explicit non-goals:** no server swing cooldown, no hold-move walk animation, no controls remapping
UI, no new bot drag scenario.

### v28 — Full equipment and belt hotbar

**Proves:** The single weapon slot can be replaced by server-authoritative paper-doll equipment while
keeping replay, persistence, bots, and the Godot UI in sync.

- Wire `equipped` now exposes `head`, `amulet`, `chest`, `gloves`, `belt`, `boots`,
  `ring_left`, `ring_right`, `main_hand`, and `off_hand`; legacy `weapon` was migrated to
  `main_hand` across schemas, fixtures, bots, smoke, and client code.
- Go sim enforces slot compatibility, logical ring slots, one-hand plus shield coexistence,
  two-handed sword/bow occupancy, and offhand blocking when `main_hand` holds a two-handed item.
- `use_hotbar_intent { slot_index }` resolves the assigned item server-side, while direct
  `use_intent { item_instance_id }` remains valid for bag use.
- Character hotbar layout persists in Postgres, session-start hotbar snapshots preserve replay
  determinism, and stale item removal clears every referencing hotbar slot.
- Base hotbar capacity is 2; belts roll `hotbar_slots` and expand capacity up to 10. Disabled slots
  retain assignments, no-op client-side when pressed, and reject server-side if explicitly used.
- `equipment_lab`, `equipment_lab_tc_1`, and `shared/golden/full_equipment.json` cover every v28
  equipment category, shield display rolls, belt capacity, and hotbar re-enable behavior.
- Godot inventory now renders named paper-doll slots and sends protocol-backed equip/unequip/hotbar
  intents; the consumable bar is snapshot/delta driven and updates capacity from authoritative
  equipment deltas.
- Protocol bot scenario `19_full_equipment.json` proves full slot coverage, hand occupancy, pinned
  belt capacity 10, disabled-slot persistence, reconnect/replay, and fresh-session persistence.
- Client bot scenario `10_full_equipment.json` proves named loot pickup, paper-doll equip, disabled
  hotbar assignment, belt expansion, and enabled hotbar use through the Godot UI path.

**Explicit non-goals:** armor mitigation, block chance execution, affix grammar, comparison UI,
stash/vendors, production icons/art, offhand abilities/dual-wield, and deeper dungeon drop economy.

### v29 — Dungeon equipment drop expansion

**Proves:** Real generated dungeon monsters and guarded chests can use the expanded v28 equipment
catalog through deterministic, depth-aware treasure classes.

- `shared/rules/dungeon_generation.v0.json` now declares temporary coarse loot bands for depth
  `1`, `2`, and `3+`; level `0` town still does not use dungeon loot bands.
- Depth-specific monster and guarded-chest loot tables bridge to new treasure classes, with chest
  equipment odds intentionally better than normal monster odds.
- Generated dungeon monsters and chests store their selected loot table at generation time, while
  source kill/open still owns all reward rolls in the Go sim.
- By depth `3+`, validation proves the configured dungeon/chest reward set can reach every v28
  equipment template: weapons, shield, armor pieces, belt, boots, ring, and amulet.
- `shared/golden/dungeon_equipment_drops.json` pins representative depth/source selection and
  monster/chest outcomes; `treasure_class_rolls.json` now covers varied direct equipment,
  potion, and money-like rolls.
- Protocol bot scenario `20_dungeon_equipment_drops.json` descends into generated dungeon play,
  opens a depth-band chest, picks up rolled equipment, equips it, and proves `/state`, reconnect,
  replay, and fresh-session persistence.

**Explicit non-goals:** final depth economy, item-level gates, Magic Find, affixes, unique/set
items, real gold wallet, vendors/stash/crafting/trade, combat use of armor/block/crit/hit speed,
production item/chest art, and client-side loot logic.

### v30 — Monster rarity and loot scaling

**Proves:** Generated dungeon monster population can roll server-authoritative rarity that changes
challenge, XP, loot depth, protocol state, replay, bot assertions, and Godot presentation.

- `shared/rules/dungeon_generation.v0.json` now declares generated monster rarities:
  `common`, `champion`, `rare`, and `unique`, with weights `100/15/6/3`, pastel colors, challenge
  multipliers, and loot-depth offsets `+0/+1/+2/+3`.
- Shared/golden validation pins rarity tuning, scaled `dungeon_mob` HP/damage/XP, seeded generated
  roll order, and the unique `level -5 -> effective depth 8 -> 3+ loot band` case.
- Go generation rolls rarity from a separate deterministic rarity RNG stream so existing floor and
  chest layout streams do not drift.
- Generated monsters store rarity, scaled HP, scaled proactive attack damage, scaled XP reward, and
  a monster loot table selected from `abs(level) + loot_depth_offset`.
- Static/lab/world-preset monsters remain unscaled and do not emit v30 generated rarity.
- Existing protocol v1 entity `rarity` now carries monster rarity through snapshots, deltas,
  `/state`, reconnect, and replay timelines.
- Godot keeps server authority unchanged and applies a green player tint plus rarity tints on the
  existing monster model: pastel white, blue, red, and golden.
- Protocol bot scenario `21_monster_rarity_loot_scaling.json` descends into a generated dungeon,
  observes a champion mob, kills it, picks up rolled loot, and proves `/state`, reconnect, replay,
  and fresh-session persistence.
- Existing character leveling bot coverage now pins a generated-dungeon seed because v30-scaled XP
  changes the expected XP total from generated mobs.

**Explicit non-goals:** unique/set item catalogs, unique monster special drops, affixes, named
elite packs, minions, aura modifiers, boss floors, Magic Find, final item-level/depth economy,
chest rarity, production monster art/VFX/audio, and colorblind/accessibility-safe rarity treatment.

### v31 — Combat stat effects and feedback

**Proves:** Player and monster combat stats now drive deterministic authoritative combat outcomes
and Godot renders those outcomes from server event metadata.

- Shared combat rules now define base hit/crit values, minimum non-blocked damage, and the global
  `75%` block cap.
- Monster rules support hit chance, crit chance, crit damage, armor, and block chance, with explicit
  combat-lab targets for miss, crit, armor-floor, block, and monster-side proofs.
- Go combat uses one deterministic resolution path for melee, projectiles, proactive monster
  attacks, and retaliation: hit roll, block roll, damage roll, crit roll, armor mitigation, and
  minimum damage.
- Misses and blocks emit combat events but do not mutate HP, trigger retaliation, kill entities,
  drop loot, or award XP; successful non-blocked hits always deal at least `1`.
- Protocol v1 combat events now expose source/target ids, outcome, raw and mitigated damage,
  `blocked`, and `critical`; progression snapshots/deltas expose effective stat breakdown rows.
- Equipped base stats, rolled equipment stats, derived character formulas, caps, and clamps are
  visible through server-owned stat breakdowns and the Godot character stats panel.
- Godot floating combat text now renders normal damage, crits, misses, and blocks from authoritative
  events, with a persisted settings toggle to suppress the presentation only.
- Protocol scenario `22_combat_stat_effects.json` and client scenario `11_combat_feedback.json`
  prove the complete path through `/state`, reconnect, replay, and headless Godot presentation.

**Explicit non-goals:** attack-speed gameplay, movement-speed gameplay, spells, mana consumers,
status effects, affix grammar, polished comparison UI, enemy equipment inventories, production
combat VFX/audio, and Protobuf migration.

### v32 — Test floor and resilient scenarios

**Proves:** CI distinguishes intentional contract locks from mutable tuning details, so normal
dungeon, population, movement, loot-weight, and presentation tuning can proceed without weakening
replay, schema, formula, persistence, or protocol coverage.

- `CLAUDE.md` now documents the durable Test Locking Policy for future slices.
- The v32 plan audit record classifies exact assertions as contract locks, behavior proofs, or
  tuning details before changing tests.
- Python bot assertions now support semantic entity filters, range comparators, inventory filters,
  eventual assertions, and generated dungeon walk budgets derived from map size.
- Protocol scenarios now prove chase/leash/dungeon behavior through eventual or semantic assertions
  instead of fixed tick waits and incidental total population counts.
- Character leveling keeps formula, level, max-HP, stat allocation, event, replay, reconnect, and
  persistence locks while avoiding an exact generated-XP tuning total.
- Go, shared validation, and GDScript golden tests keep schema and formula contracts exact while
  deriving or structurally validating generated population, rarity tuning, and loot-depth offsets.
- Client bot scenarios can target entities by debug metadata such as monster definition,
  interactable definition, item definition, rarity, and state instead of fragile entity indexes.
- Local reverted probes changed dungeon floor size, generated monster population, and movement speed;
  the focused checks stayed green with no committed tuning changes.

**Explicit non-goals:** no gameplay, balance, protocol, or UI feature work; no committed tuning
changes; no broad test framework migration; no `tuning_sensitive` metadata.

### v33 — True co-op session

**Proves:** Two authenticated clients can join one Go-authoritative session, each controlling a
distinct character/player entity while replay, persistence, reconnect, and solo compatibility stay
deterministic.

- Protocol v2 schemas add `local_player_id`, `party[]`, actor metadata on player/combat/reward
  events, and actor-free client intents; v1 schemas remain intact.
- Sessions now support `mode: "coop"`, hashed join codes, deterministic `session_members`,
  per-member start snapshots, and actor-tagged input rows.
- HTTP create/join lets a host create a co-op session and a guest join by session id + join code;
  non-members are denied WebSocket access and duplicate member sockets are rejected.
- The sim now owns multiple player states with independent levels, inventories, hotbars,
  waypoints, progression, reconnect-to-town behavior, and non-solid player/player collision.
- A shared realtime session loop runs one authoritative sim per active session, binds each socket
  to its server-derived actor, sends recipient-scoped snapshots, and fans out level-visible deltas.
- Disconnecting a co-op member removes only that player's entity from other same-level clients;
  solo disconnects continue to preserve same-session resume behavior.
- Replay reconstructs host and late-joined guest members with actor-tagged inputs, member start
  snapshots, disconnect/reconnect state, and per-tick event sequence ordering.
- Godot stores `local_player_id`, keeps the local `PlayerAnchor` for camera/prediction/input, and
  renders other visible players as remote entity nodes with authoritative movement only.
- Protocol bot scenario `23_true_coop_session.json` proves host create, guest join, distinct local
  player ids, party metadata, same-level visibility, independent movement, guest disconnect/removal,
  guest reconnect to town, and replay verification.

**Explicit non-goals:** matchmaking/lobby, public discovery, Steam lobby/invites, party panel
polish, chat/emotes/ready checks, trade, XP sharing, party bonuses, loot allocation rules,
friendly fire/PvP, production remote-player art, more than two players, and distributed session
ownership across server processes.

### v34 — Model reaction polish

**Proves:** Damage/death presentation can be improved for all character-like visible entities while
staying client-only and driven by existing authoritative combat events.

- Godot now attaches a `ModelReactionController` to the local player, remote co-op players, and
  monsters, layering transform/material reactions over the existing `AnimationController`.
- Hit reactions lean away from the resolved attacker when possible, briefly dark-blink the model,
  then restore the entity's base tint.
- Death reactions supersede active hit/locomotion presentation, rotate the model down, and leave a
  persistent darker corpse presentation.
- Snapshot/render paths apply terminal death presentation from `hp <= 0`, so already-dead monsters
  and players do not need the original kill event to look dead.
- Remote co-op players now instantiate the same humanoid character scene as the local player and
  use a readable dark charcoal tint while remaining server-authoritative and unpredicted.
- Client bot debug state exposes local/entity presentation metadata for headless assertions without
  pixel matching.
- Client scenario `12_model_reaction_polish.json` proves monster hit, local-player hit from
  retaliation, and monster terminal death presentation through the real Godot client.

**Explicit non-goals:** no server/protocol/schema changes, no external animation plugin, no
production customization/cosmetics, no monster art replacement, no corpse collision/despawn, and no
respawn/revive behavior.

### v35 — Boss floor gate

**Proves:** The generated dungeon can introduce a compact skill-gated boss floor with locked exits,
telegraphed damage, boss presentation metadata, and deterministic replay coverage.

- Dungeon level `-5` is generated as a compact `30 x 30` boss floor with fixed up/down stairs,
  one disabled teleporter, one pre-boss chest, one humanoid boss, and reduced trash population.
- Boss floor down stairs and teleporter start locked/disabled with reason `boss_alive`; both
  transition to `ready` and emit state-change events when the boss dies.
- Shared boss template and pattern rules define the first humanoid `cave_warden` boss, visual
  model/tint/scale metadata, and a telegraphed charged melee pattern with active-only damage.
- Go sim owns boss phase timing, hit predicates, locked-exit rejection, unlock, level transition,
  boss loot/chest hooks, and deterministic replay reconstruction.
- Protocol schemas now carry optional boss/visual metadata, boss phase events, lock/unlock events,
  and disabled/locked interactable state without requiring a client intent shape change.
- Godot renders boss entities through the humanoid model path, applies authoritative visual scale
  and tint, shows telegraph tinting, and presents locked/ready exit state from server data.
- Protocol bot scenario `24_boss_floor_gate.json` proves descent to `-5`, compact floor metadata,
  locked exit rejects, boss phase observation, boss kill, exit unlock, descent to `-6`, `/state`,
  reconnect, and replay verification.

**Explicit non-goals:** no additional boss templates, enrage phases, summoned adds, production
boss art/VFX/audio, boss health bar UI, co-op boss scaling, durable boss/map snapshots, quest
integration, or gameplay collision/reach scaling from visual scale.

### v36 — Inventory paper-doll capacity

**Proves:** Inventory capacity can be server-owned, item-derived, and rendered as a fixed
paper-doll bag grid without making the Godot UI authoritative.

- Shared item template rules now allow `inventory_rows` and include deterministic
  `cave_pack_belt`, with `shared/golden/inventory_capacity.json` pinning base rows `3`,
  5-column base capacity `15`, +1 row capacity `20`, hotbar/equipped exemptions, and rejection
  reasons.
- Session snapshots expose `inventory_rows` and `inventory_capacity`; relevant equipped/hotbar
  deltas publish the same fields so client, bot, reconnect, and replay observe identical capacity.
- Go sim derives capacity from equipped items, counts only bag entries that are not equipped and not
  assigned to hotbar, rejects full pickups with `inventory_full`, and rejects capacity-shrinking
  unequip/unassign paths before mutation with `capacity_would_overflow`.
- Godot replaces the two-column equipment list with a named paper-doll layout around a
  `character_paper_doll` placeholder, renders a 5-column bag with exactly `inventory_capacity`
  visible cells, and keeps the bag drop target outside grid math.
- Inventory debug state now reports paper-doll slot ids/positions, preview status, capacity rows,
  bag columns, available slot count, and empty-slot style markers for headless assertions.
- Protocol bot scenario `25_inventory_capacity_and_paper_doll.json` proves base capacity, full-bag
  rejection, +1 row equip to capacity 20, five more bag entries, reconnect, and replay.
- Client bot scenario `13_inventory_paper_doll.json` proves base 15-cell grid, all paper-doll slot
  ids, the preview node, belt equip, and expanded 20-cell grid.

**Explicit non-goals:** no stash, vendors, crafting, item sorting/filtering, comparison UI,
multi-cell item footprints, passive skill sources for inventory rows, production paper-doll art,
or full model-backed SubViewport preview.

---

## Architecture decisions (ADRs)

| ADR | Topic | Status |
|-----|-------|--------|
| [0001](adr/0001-technology-stack.md) | Foundational stack (Go server, Godot client, shared rules, replay, bot) | Accepted |
| [0006](adr/0006-asset-pipeline.md) | glTF-first assets, manifests, sockets, validation | Accepted; v3 as-built for rigged GLBs |
| [0007](adr/0007-animation-state-model.md) | Client-only animation; event-driven reactions | Accepted; v4 as-built for player reactions |
| [0008](adr/0008-world-structure-and-dungeon-progression.md) | World structure: infinite inverted-tower dungeon, multi-level Sim, character-scoped persistence, waypoints, co-op | Accepted |

Anticipated but **not written:** netcode timing, Protobuf migration, production auth, multiplayer split,
quest system design, NPC interaction protocol, player trade, character progression formulas
(see ADR-0001 follow-up list and ADR-0008 deferred items).

---

## Scripted vertical slice flow (bot + smoke)

Every slice keeps this loop working unless the spec explicitly changes it:

```text
dev-login → create session → move → attack training dummy → pick up loot → equip rusty_sword
```

After v4 the player **survives with reduced HP** (`hp < 10`). Monster dies; player may take retaliation
each successful hit. After v7 this flow lives in `tools/bot/scenarios/01_vertical_slice.json`; additional
scenario JSON files are automatically included in filename order in `make bot` and `make bot-visual`.

The scenario catalog also includes:

```text
gear_before_combat: walk to rusty_sword → pick up → equip → one-shot reward dummy → pick up training_badge
collision_lab: pass through middle wall gap → kill monster on far side
inventory_lab: pick up rusty_sword → equip → unequip → drop → re-pickup → re-equip
heal_lab: pick up red_potion x2 → take damage → use potion twice → full HP
chase_lab / chase_maze / leash_lab: wait while chase monster closes; kite beyond leash and return
dungeon_levels / teleporter_lab: start in town, descend/ascend generated floors; discover teleporters and fast-travel back
character_persistence: same-account fresh sessions retain gear/equipment and discovered waypoint access
rolled_drops: kill dungeon mob → pick up/equip rolled cave_blade → prove rolled metadata persists
main_menu_flow: menu settings → named character creation → pause input lock → return → continue fresh session
treasure_classes_and_guarded_chests: pinned chest floor → kill guarded mob → open chest once → pick up chest loot
character_stats_and_leveling: descend to dungeon → kill mobs for XP → level up → spend VIT → prove persistence
full_equipment: pick up/equip paper-doll gear → prove hand occupancy → assign belt-gated hotbar → prove persistence
dungeon_equipment_drops: descend to depth-banded dungeon → open chest → pick up/equip rolled equipment → prove persistence
monster_rarity_loot_scaling: descend to generated dungeon → assert champion rarity → kill → pick up rolled loot → prove persistence
combat_stat_effects: combat lab proofs for miss, crit, armor floor, block, monster crit/block, projectile impact, and stat breakdowns
client_combat_feedback: equip gear → assert stat breakdowns → prove normal/crit/miss/block floating text and settings toggle
true_coop_session: host creates co-op → guest joins → shared-level visibility → independent movement → disconnect/reconnect → replay proof
model_reaction_polish: attack training dummy → prove monster hit reaction → prove local player hit reaction → kill dummy → prove terminal corpse reaction
boss_floor_gate: descend to level -5 → assert compact boss floor and locked exits → observe boss phase → kill boss → unlock exits → descend to -6
inventory_capacity_and_paper_doll: fill base 15-capacity bag → reject full pickup → equip capacity belt → fill expanded 20-capacity bag
```

**Verify:**

```bash
make db-up && make server    # terminal 1
make bot                     # terminal 2 — all protocol bot scenarios
make client-unit             # headless Godot unit gates (no server required)
make client-smoke            # headless Godot gates + slice smoke
make bot-client              # Godot client bot scenarios; requires live server
make ci                      # full suite
make bot-visual              # optional — record all bot scenarios and watch replay playlist in Godot
make bot-visual scenario=07_inventory_lab.json  # optional — replay one scenario by file name or id
```

---

## Open gaps & deferred work

Do **not** assume these are the next slice — they are documented backlog items agents should know about.

### Recently closed

**Combat/world state now persists on same-session resume.** v5 replays recorded
inputs before the WebSocket `session_snapshot`, so monster death, player HP,
inventory, equipped state, and ID continuity are restored authoritatively.

**World preset identity now persists on sessions.** v7 stores `world_id`, so fresh WebSocket attach,
resume, `/state`, replay verification, and replay timeline all reconstruct the same initial layout.

**Equipped weapon damage now changes authoritative combat.** v8 resolves `rusty_sword.damage`
from equipped server state at attack time and proves the equipped gear scenario kills the reward
dummy in one acknowledged attack.

**Solid collision now blocks movement through bodies and walls.** v9 resolves player movement
against live monsters and static world walls, while collision lab proves routed movement and
deterministic replay.

**Click action and melee reach are now authoritative.** v10 unifies combat/pickup/door activation
behind `action_intent`, enforces reach from shared rules, and proves a replayable opening door.

**Click-to-move and action auto-approach are now authoritative.** v11 adds shared navigation
rules, deterministic server A*, `move_to_intent`, and a path-maze bot proof.

**Ranged projectile combat is now authoritative.** v12 adds ranged weapon rules, projectile
entities, swept collision, impact-time hit/damage, and a ranged-lab bot/replay proof.

**Inventory UI, unequip, and player drop are now authoritative.** v13 adds protocol-backed
unequip/drop intents, deterministic adjacent loot placement, persisted inventory removal, and a
display-only Godot panel that mirrors server snapshots/deltas.

**Current item presentation is now shared-data-driven.** v15 adds presentation metadata for all
current item definitions and uses it for inventory icons and ground loot silhouettes without
server/protocol changes.

**Consumable healing is now authoritative.** v16 adds `use_intent`, red potion heal rules, HP cap
goldens, server-owned inventory removal, and a client-only hotbar that sends use intents.

**Monster chase movement is now authoritative.** v17 adds opt-in chase behavior, deterministic
monster pathing around solids, leash return, chase/lab bot scenarios, and client walk presentation
from position deltas.

**Dungeon levels, stairs, teleporters, town entry, and dungeon monster threat are now authoritative.**
v18 adds multi-level dungeon state and generated stairs; v19 adds generated teleporters, session
discovery, and server-owned fast travel with a client-only waypoint panel; v20 makes town level `0`
the fresh play-session entry and keeps dungeon floors lazy; v21 spawns deterministic hostile dungeon
mobs that chase and proactively damage the player.

**Character inventory/equipment and waypoint unlocks now persist across fresh sessions.** v22 moves
durable item instances and discovered waypoint levels to the default character, while preserving
session-start snapshots for deterministic replay and keeping HP, dungeon maps, monsters, corpses,
opened doors, and floor drops session-scoped.

**Dungeon mobs now drop rolled weapon gear.** v23 adds server-authoritative item templates,
deterministic rarity/stat rolls, rolled weapon damage, rolled payload persistence, and tooltip
presentation for the first rolled weapon template.

**The client now has a player-facing menu shell.** v24 adds named character list/create APIs,
fresh-session Continue/New Game flows, local window-size settings, ESC pause, Return to Main Menu,
and a Godot client bot proof for the complete menu path.

**Treasure classes and guarded chests are now authoritative.** v25 adds data-driven multi-attempt
monster/chest loot, deterministic rare chest generation with guarded monster bonus, open-once chest
loot, and bot/golden coverage for the complete path.

**Character stats and leveling are now authoritative.** v26 adds durable XP, levels, stat points,
base stats, derived substats, stat allocation, VIT max-HP effects, STR damage contribution, a
Godot character sheet, an XP bar, and protocol/client bot proofs.

**Sustained left-click controls are now client-side.** v27 adds hold-to-attack on monsters and
hold-to-move on floor by repeating existing `action_intent` / `move_to_intent` at `SEND_INTERVAL`,
with sticky targets, move epsilon, and headless unit coverage — no protocol or server changes.

**Full paper-doll equipment and belt-gated hotbar are now authoritative.** v28 replaces the single
weapon slot with full equipment slots, two-hand occupancy, droppable gear templates, persisted
character hotbar layout, replay-safe session hotbar snapshots, and protocol/client bot proofs for
server-synced paper-doll and belt capacity behavior.

**Generated dungeon drops now reach the expanded equipment catalog.** v29 adds temporary depth
bands, depth-specific monster/chest treasure classes, validation for full v28 template reachability
by depth `3+`, golden fixtures for varied equipment outcomes, and a real generated dungeon bot proof.

**Generated dungeon monster rarity now scales challenge and loot depth.** v30 adds deterministic
generated monster rarity tiers, scaled HP/damage/XP, effective monster loot depth offsets,
monster rarity in protocol/replay state, player/enemy tinting, and a real generated dungeon bot
proof for non-common rarity.

**Combat stats now affect authoritative outcomes.** v31 applies hit, crit, armor, block, minimum
damage, and effective stat breakdowns across player and monster combat, then renders normal, crit,
miss, and block feedback from protocol events in Godot.

**The test floor now separates contracts from tuning details.** v32 keeps exact locks for replay,
schema, formula parity, persistence boundaries, and named UI/protocol contracts, while converting
brittle dungeon size, generated population, movement timing, rarity tuning, and selector-index
assumptions to semantic, range, derived, or eventual checks.

**True two-player co-op sessions are now authoritative.** v33 adds server-owned co-op session
membership, hashed join codes, actor-tagged inputs, per-player sim state, recipient-scoped realtime
snapshots/deltas, remote-player Godot rendering, and a protocol bot proof for join, movement,
disconnect/reconnect, and replay.

**Character-like model reactions are now unified in the Godot client.** v34 adds client-only
hit/death transform and tint reactions for local players, remote co-op players, and monsters;
remote co-op players now reuse the humanoid character model with a distinct dark tint.

**The first boss floor gate is now authoritative.** v35 adds a compact level `-5` boss arena,
telegraphed boss phases, locked down-stair/teleporter exits until boss death, boss visual scale
metadata, and protocol/replay/bot proof for unlock and descent.

**Inventory capacity and the paper-doll bag grid are now authoritative.** v36 adds server-derived
`inventory_rows` / `inventory_capacity`, an item-granted row source, full-bag and overflow rejection
guards, a 5-column capacity grid, and protocol/client bot proofs.

### Other deferred items (from specs / ADRs)

| Area | Deferred item | Source |
|------|---------------|--------|
| Persistence | Player-facing old-session resume, delete/rename characters, class selection, visual customization, portraits, main-menu character summaries, stash/vendors/gold, quest progress, passive skills, respec, respawn/checkpoints, durable dungeon map snapshots | v22/v24/v26 non-goals, ADR-0008 deferred |
| Combat | Attack-speed gameplay, mana consumers/regeneration, respawn, spell systems, piercing/AoE/homing projectiles, ranged monster AI, depth scaling beyond loot bands, offhand abilities/dual-wield, named elite packs/minions/aura modifiers, additional boss templates/pattern decks, enrage phases, summoned adds, co-op boss scaling | v0/v4/v12/v17/v21/v23/v26/v28/v29/v30/v31/v32/v35 non-goals |
| Itemization | Affix grammar, procedural item names, stat requirements, special-effect execution, comparison UI, loot filters, crafting/vendors/gold/trade, real gold wallet, Magic Find, unique/set catalogs, unique monster special drops, final item-level/depth progression, richer boss drop economy, richer dungeon drop economy, item sorting/filtering, multi-cell item footprints, passive skill sources for inventory rows | v23/v25/v26/v28/v29/v30/v35/v36 non-goals, ADR-0009 deferred |
| Content | Production item art/icons, production menu art/audio, production town art, production chest art/animation/audio, production monster art/VFX/audio, production boss art/VFX/audio, production combat VFX/audio, production paper-doll art/model preview, colorblind/accessibility-safe rarity presentation, NPCs/vendors/stash, additional item families beyond current rules | v15/v20/v23/v24/v25/v28/v29/v30/v31/v32/v35/v36 non-goals |
| Settings | Fullscreen, audio, controls remapping, accessibility options, graphics quality, language selection | v24 non-goals |
| Assets | Blender export pipeline, texture budget, remote patcher | ADR-0006 |
| Platform | Production auth provider, dashboards, historical inspect API | v0 §8, ADR-0001 |
| Protocol | Protobuf / `godobuf` migration | ADR-0001 |
| Multiplayer | Matchmaking/lobby, public session discovery, Steam lobby/invites, friend flows, party UI polish, chat/emotes/ready checks, trade, XP sharing, party bonus, proximity reward rules, loot allocation, friendly fire/PvP, production remote-player art, more than two players, split deployables / cross-process session ownership | v0/v33 non-goals, ADR-0001 |

---

## Starting a new task (agent checklist)

1. **Read this file** (`docs/PROGRESS.md`) — confirm baseline slice and open gaps.
2. **Read ADR-0001** and any feature-specific ADRs listed above.
3. **Spec first** — create or read `docs/specs/vN_spec-<feature>.md` (SDD; `N` = next execution order).
4. **Plan second** — create `docs/plans/vN_<YYYY-MM-DD>-<feature>.md` with file map + verification commands.
5. **Branch** — stay on the current checkout; do not create branches (user creates them before development if needed).
6. **Implement** shared → server → client → bot/smoke → docs; keep `make ci` green.
7. **Update this file** when the slice completes: new row in lifecycle table, summary, and any new gaps.

### Invariants (do not break)

- Go sim determinism: seeded RNG only, no wall-clock in `game/`, stable ordering.
- Shared rules are **data**; formulas evaluated in Go + GDScript from the same golden fixtures.
- Animation is client-only; new reactions need a **server event** first, then client mapping.
- Golden changes require Go tests **and** GDScript `test_golden.gd` / `validate_shared.py` updates.

---

## Repo map (quick reference)

```text
client/          Godot 4.6.3 — main.gd, animation_controller.gd, net_client.gd, smoke.gd
server/          Go — internal/game (sim), internal/realtime (WS), internal/store (Postgres)
shared/          protocol schemas, rules JSON, golden fixtures
tools/           bot, replay, validate_shared.py, assets/
assets/          manifests + gen scripts
docs/            ADRs, specs, plans, this file
```

**Agent entrypoints:** [`CLAUDE.md`](../CLAUDE.md) (commands + architecture), this file (progress),
[`README.md`](../README.md) (human onboarding).

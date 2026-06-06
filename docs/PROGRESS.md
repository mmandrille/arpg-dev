# Project progress & slice lifecycle

**Read this file at the start of every new task** before writing specs, plans, or code.
It is the canonical snapshot of what exists, what each slice proved, and what is still open.

Last updated: 2026-06-05

---

## Current status

| Field | Value |
|-------|-------|
| **Latest completed slice** | v13 — `inventory-ui` (authoritative inventory panel, unequip, and drop) |
| **Active branch** | `feature/solid-collision-and-obstacles` |
| **CI gate** | `make ci` green on 2026-06-05 |
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

---

## Architecture decisions (ADRs)

| ADR | Topic | Status |
|-----|-------|--------|
| [0001](adr/0001-technology-stack.md) | Foundational stack (Go server, Godot client, shared rules, replay, bot) | Accepted |
| [0006](adr/0006-asset-pipeline.md) | glTF-first assets, manifests, sockets, validation | Accepted; v3 as-built for rigged GLBs |
| [0007](adr/0007-animation-state-model.md) | Client-only animation; event-driven reactions | Accepted; v4 as-built for player reactions |

Anticipated but **not written:** netcode timing, Protobuf migration, production auth, multiplayer split
(see ADR-0001 follow-up list).

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
```

**Verify:**

```bash
make db-up && make server    # terminal 1
make bot                     # terminal 2 — all protocol bot scenarios
make client-smoke            # headless Godot gates + slice smoke
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

### Other deferred items (from specs / ADRs)

| Area | Deferred item | Source |
|------|---------------|--------|
| Persistence | Cross-session **character-scoped** inventory | v0 as-built §10 |
| Combat | Healing, armor, respawn, spell systems, piercing/AoE/homing projectiles, monster ranged AI | v0/v4/v12 non-goals |
| Content | Visual mappings for items beyond `rusty_sword` | equip spec §4.9 |
| Assets | Blender export pipeline, texture budget, remote patcher | ADR-0006 |
| Platform | Production auth provider, dashboards, historical inspect API | v0 §8, ADR-0001 |
| Protocol | Protobuf / `godobuf` migration | ADR-0001 |
| Multiplayer | Matchmaking, multi-player sessions, split deployables | v0 non-goals, ADR-0001 |

---

## Starting a new task (agent checklist)

1. **Read this file** (`docs/PROGRESS.md`) — confirm baseline slice and open gaps.
2. **Read ADR-0001** and any feature-specific ADRs listed above.
3. **Spec first** — create or read `docs/specs/vN_spec-<feature>.md` (SDD; `N` = next execution order).
4. **Plan second** — create `docs/plans/vN_<YYYY-MM-DD>-<feature>.md` with file map + verification commands.
5. **Branch** — `feature/<codename>` off latest integration branch (today: merge target TBD).
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

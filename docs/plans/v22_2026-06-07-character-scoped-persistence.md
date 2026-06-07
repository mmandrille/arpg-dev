# v22 Plan — Character-scoped persistence

Status: Ready for implementation
Goal: Make default-character item instances, equipped weapon state, and waypoint unlocks survive fresh sessions for the same authenticated account.
Architecture: Item instances and waypoint progress move from session-scoped persistence to character-scoped persistence. Fresh sessions load the character's durable progression into a new authoritative Sim, while replay and same-session resume reconstruct from immutable session-start snapshots plus recorded inputs. Dungeon layout, monsters, HP, floor drops, and session event/input logs remain session-owned. The item table is future-ready for randomly rolled per-instance stats and later stash ownership.
Tech stack: Go HTTP/realtime/replay/store, Postgres migrations, shared protocol snapshots already in place, Python protocol bot, Godot client as existing presentation only.

## Baseline and shortcut decision

v22 builds on v21 `dungeon-monster-combat`: fresh play starts in town level `0`, dungeon levels are generated from a fresh session seed, dungeon mobs threaten the player, and existing inventory/waypoint UI already renders snapshot and delta state.

Godot shortcut adoption checklist:

- **Decision:** reject plugin adoption.
- **Reason:** this slice has no new Godot UI, camera, placeholder art, inventory presentation, or isometric tooling. Existing inventory and waypoint panels should render persisted snapshot state without client feature work.
- **Borrow:** existing client-unit/smoke coverage only if server snapshot shape changes unexpectedly.

Spec defaults locked for implementation:

- Preserve existing `inventory_items` rows where practical, but prefer clean character item-instance ownership over compatibility-only complexity.
- Model owned items as durable instances with `location` and `rolled_stats`, even though v22 only exercises inventory/equipped locations and does not roll stats.
- Add immutable session-start progression snapshots so replay and visual timeline do not read mutable live character state.
- Persist only activated/discovered teleporters, not merely generated or visited floors.
- Keep player HP session-scoped.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/migrations/0001_init.sql` | Update comments only if needed to reflect v22 ownership |
| Add | `server/migrations/0003_character_progression.sql` | Add character item-instance, waypoint, and session-start snapshot tables |
| Modify | `server/internal/store/models.go` | Add character item-instance, waypoint, and session-start snapshot models |
| Modify | `server/internal/store/interfaces.go` | Replace session inventory repo methods with character item/session-start progression methods |
| Modify | `server/internal/store/repos.go` | Implement character item, waypoint, and snapshot persistence |
| Modify | `server/internal/store/store_test.go` | Cover durable item/waypoint repos and snapshot isolation |
| Modify | `server/internal/game/sim.go` | Add helpers to load discovered teleporters and export discovery changes if missing |
| Modify | `server/internal/http/session.go` | Create fresh sessions with default-character progression snapshot |
| Modify | `server/internal/realtime/hub.go` | Load session-start progression for fresh attach and replay boundary |
| Modify | `server/internal/realtime/runner.go` | Persist inventory and waypoint mutations by character and keep session events |
| Modify | `server/internal/replay/replay.go` | Reconstruct and build timelines from session-start progression snapshot plus inputs |
| Modify | `server/internal/http/ws_test.go` | Cross-session inventory/equipment/waypoint and replay tests |
| Modify | `server/internal/http/auth_session_test.go` | Account isolation and create/resume regression tests |
| Modify | `tools/bot/run.py` | Add multi-session same-account scenario support and assertions |
| Add | `tools/bot/scenarios/15_character_persistence.json` | End-to-end durable progression proof |
| Modify | `docs/PROGRESS.md` | Add v22 lifecycle summary when complete |

## Task 1 — Database schema and store contracts

Files:

- Add: `server/migrations/0003_character_progression.sql`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/store/store_test.go`
- Modify: `server/migrations/0001_init.sql` only for stale comments if needed

- [x] Step 1.1: Add `character_item_instances` keyed by a durable item instance id with `character_id`, `account_id`, `item_def_id`, `location`, `slot`, `equipped`, `rolled_stats JSONB`, and timestamps.
- [x] Step 1.2: Add `character_waypoints(character_id, level, discovered_at)` with `PRIMARY KEY(character_id, level)`.
- [x] Step 1.3: Add immutable session-start snapshot tables, preferably `session_start_item_instances` and `session_start_waypoints`, keyed by `session_id`, to freeze the starting character state for replay and visual timelines.
- [x] Step 1.4: Migrate existing `inventory_items` into character-owned item instances where possible using `session.character_id`; tolerate duplicate per-session deterministic item IDs by minting a durable character item id while preserving the protocol `item_instance_id` needed by Sim/replay snapshots.
- [x] Step 1.5: Add store models for character item-instance rows, including `Location` and raw `RolledStats`, plus waypoints and session-start snapshots.
- [x] Step 1.6: Add repo methods:
      `ListCharacterItems`, `AddCharacterItem`, `SetCharacterItemLocation`, `SetCharacterItemEquipped`, `RemoveCharacterItem`,
      `ListCharacterWaypoints`, `AddCharacterWaypoint`,
      `CreateSessionStartSnapshot`, and `LoadSessionStartSnapshot`.
- [x] Step 1.7: Keep inventory mutation methods idempotent for replay/same input retry, but do not silently cross account/character ownership boundaries.
- [x] Step 1.8: Add store tests for cross-session load, equipped state, use/drop removal, `rolled_stats` round-trip, waypoint insert idempotency, account isolation, and session-start snapshot immutability.

```bash
cd server && go test ./internal/store
```

## Task 2 — Sim progression load helpers

Files:

- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go` if a small exported load type is needed
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Reuse `LoadInventory` for fresh character item instances and session-start item snapshots; ensure it advances `nextID` past loaded numeric item IDs to avoid collision with newly spawned loot pickups in the same session.
- [x] Step 2.2: Add a `LoadDiscoveredTeleporters(levels []int)` helper that merges durable waypoint levels into `discoveredTeleporters` without removing town level `0`.
- [x] Step 2.3: Ensure loaded negative waypoint levels appear in `Snapshot().DiscoveredTeleporters` even before that generated dungeon level has been visited in the current session.
- [x] Step 2.4: Add or update game tests proving a loaded waypoint can be used from a reachable current teleporter and generates the target level from the current session seed.
- [x] Step 2.5: Preserve deterministic ordering in `teleporterDiscoveryView` by sorting levels; avoid iterating Go maps directly in observable output.

```bash
cd server && go test ./internal/game/... -run 'Inventory|Teleporter|Waypoint|Snapshot'
```

## Task 3 — Fresh session bootstrap and replay boundary

Files:

- Modify: `server/internal/http/session.go`
- Modify: `server/internal/realtime/hub.go`
- Modify: `server/internal/replay/replay.go`
- Modify: `server/internal/http/auth_session_test.go`
- Modify: `server/internal/http/ws_test.go`

- [x] Step 3.1: On fresh `POST /v0/sessions`, load the account's default character progression after creating the session.
- [x] Step 3.2: Persist an immutable session-start snapshot containing the character item instances/equipment and discovered waypoint levels for the new session.
- [x] Step 3.3: On first WebSocket attach with no recorded inputs, create `game.NewSimWithWorld`, then load the session-start inventory and waypoints before sending `session_snapshot`.
- [x] Step 3.4: On WebSocket resume with recorded inputs, continue to use `replay.Reconstruct`; do not reload live character progression.
- [x] Step 3.5: Update `replay.Reconstruct` and `BuildTimeline` to load session-start progression before applying recorded inputs or emitting the replay snapshot.
- [x] Step 3.6: Add tests proving changing live character item instances after a session starts does not change replay output for that historical session.
- [x] Step 3.7: Add tests proving another account cannot resume, inspect, or mutate a session/character it does not own.

```bash
cd server && go test ./internal/http/... ./internal/replay/... -run 'Character|Persistence|Waypoint|Replay|Session'
```

## Task 4 — Persist live item and waypoint mutations

Files:

- Modify: `server/internal/realtime/runner.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/http/ws_test.go`

- [x] Step 4.1: Change `persistTick` inventory add/update/remove handling to write durable character item instances by `sess.CharacterID`, while keeping session events and inputs session-owned.
- [x] Step 4.2: Persist `OpTeleporterDiscoveryUpdate` changes with `Discovered == true` to `character_waypoints`.
- [x] Step 4.3: Keep town level `0` always present; persist it only if doing so simplifies queries, but do not require a player action to unlock it.
- [x] Step 4.4: Treat dropped and consumed items as durable removals for v22; later stash moves should update `location = 'stash'` instead of deleting the item instance.
- [x] Step 4.5: Add integration tests for pickup/equip across fresh sessions, potion use removal, drop removal, re-pickup restoration, and waypoint persistence.
- [x] Step 4.6: Confirm old same-session reconnect behavior still reconstructs current live session state from inputs and does not duplicate durable rows.

```bash
cd server && go test ./internal/http/... -run 'Inventory|Character|Waypoint|Reconnect'
```

## Task 5 — Bot scenario

Files:

- Modify: `tools/bot/run.py`
- Add: `tools/bot/scenarios/15_character_persistence.json`
- Modify: `tools/bot/test_protocol.py` if scenario parsing/assertions change

- [x] Step 5.1: Extend the bot runner with a scenario action that creates a second fresh session using the same login token/account without re-running dev-login as a new account.
- [x] Step 5.2: Add assertions for inventory item definition present, equipped weapon definition, discovered waypoint level present, and generated current-session level after teleporting to a persisted waypoint. Keep `rolled_stats` coverage in store/HTTP tests unless the bot gains a debug-state assertion.
- [x] Step 5.3: Add scenario `15_character_persistence.json`: start `dungeon_levels`, descend, kill or avoid `dungeon_mob` as needed, pick up/equip durable gear, discover level `-1` teleporter, create a second fresh session, assert gear/equipment/waypoint in initial state, teleport to `-1`, and verify `/state`, reconnect, and replay.
- [x] Step 5.4: Include a destructive item proof where feasible: use or drop a known item, start another fresh session, and assert it is absent. If this makes the scenario too long or flaky, cover it in HTTP integration tests and document that choice in the scenario description.
- [x] Step 5.5: Keep scenarios `01`-`14` green; adjust only if character persistence changes default inventory assumptions.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
make db-up
make bot
```

## Task 6 — Client verification

Files:

- Modify: `client/tests/test_golden.gd` only if protocol snapshot shape changes
- Modify: `client/scripts/main.gd` only if existing panels fail to render persisted snapshot state

- [x] Step 6.1: Verify no new UI code is needed: persisted inventory and waypoint state should arrive through existing `session_snapshot` fields.
- [x] Step 6.2: Run client unit tests to catch schema/fixture assumptions.
- [x] Step 6.3: Run smoke if server changes affect initial snapshots or waypoint panel behavior.

```bash
make client-unit
make client-smoke
```

## Task 7 — Lifecycle docs and CI

Files:

- Modify: `docs/PROGRESS.md`

- [x] Step 7.1: When implementation ships, add v22 to the lifecycle table and mark latest completed slice as `character-scoped-persistence`.
- [x] Step 7.2: Document as-built behavior: character-owned inventory/equipment, character waypoints, session-start snapshots for replay, fresh map generation, and session-scoped HP/floor drops.
- [x] Step 7.3: Record deferred follow-ups: character picker, player-facing resume, stash/vendors/gold, quest progress, stats/skills/XP, respawn/checkpoints, and durable dungeon map snapshots.

```bash
make ci
```

## Final verification

- [x] `cd server && go test ./internal/store`
- [x] `cd server && go test ./internal/game/... -run 'Inventory|Teleporter|Waypoint|Snapshot'`
- [x] `cd server && go test ./internal/http/... ./internal/replay/... -run 'Character|Persistence|Waypoint|Replay|Session|Reconnect'`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make client-unit`
- [x] `make bot`
- [x] `make ci`

Manual check:

```bash
make play
# Pick up/equip an item and discover a dungeon teleporter.
# Restart play with the same dev account; inventory/equipment and waypoint unlocks should remain.
```

## Deferred scope

- No character select UI or multiple characters.
- No durable dungeon map, monster state, corpses, opened doors, floor drops, or player HP.
- No stash UI/interactions, vendors, gold, crafting, quests, character stats, skills, level/XP, respawn, or checkpoints.
- No random item stat generation yet; v22 only persists the future-ready `rolled_stats` payload.
- No new Godot UI, production art, or plugin adoption.

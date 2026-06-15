# v20 Plan — Play session loop

Status: Complete; `make ci` green on 2026-06-06
Goal: Turn `dungeon_levels` into the default playable loop — town at level `0`, lazy dungeon descent,
town waypoint from session start, and `make play` wired to a fresh run.

Architecture: Static town `LevelState` from `worlds.v0.json` at level `0`; negative floors remain
lazy `GenerateDungeonLevel` output. Protocol v1 relaxes teleporter level constraints to include `0`.
Tech stack: Go authoritative sim, shared JSON contracts, Godot client, Python protocol bot.

## Baseline and shortcut decision

- v18 introduced multi-level sim, stairs, scoped deltas, level HUD.
- v19 added teleporters, session discovery, waypoint panel, `teleport_intent`.
- v20 reuses all transition/discovery machinery; the work is **entry-point and town bootstrap**,
  not new protocol messages.

Godot shortcut adoption checklist:

- **Reason:** town is placeholder interactables + existing waypoint panel; no new UI surface.
- **Borrow:** existing stair/teleporter meshes, waypoint panel, level HUD hide-at-zero behavior.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/worlds.v0.json` | Town entities + player spawn on `dungeon_levels` |
| Modify | `shared/protocol/messages.v1.schema.json` | `target_level` allows `0` |
| Modify | `shared/protocol/session_snapshot.v1.schema.json` | `teleporter_discovery.level` allows `0` |
| Modify | `shared/protocol/state_delta.v1.schema.json` | discovery update `level` allows `0` |
| Modify | `shared/golden/dungeon_stairs.json` | Town round-trip case for replay/golden tests |
| Modify | `shared/golden/dungeon_stairs.v0.schema.json` | Allow town level `0` in the fixture |
| Modify | `shared/golden/dungeon_teleporters.json` | Include town row in discovery fixtures |
| Modify | `server/internal/game/dungeon_gen.go` | Add level `-1` up stair for town descent |
| Modify | `server/internal/game/sim.go` | Town bootstrap, travel to/from level `0`, discovery init |
| Modify | `server/internal/game/game_test.go` | Town start, ascend to town, teleport to town, deep descend |
| Modify | `server/internal/game/game_replay_test.go` | Replay teleporter golden starts from town |
| Modify | `tools/bot/scenarios/12_dungeon_levels.json` | Preamble: descend from town |
| Modify | `tools/bot/scenarios/13_teleporter_lab.json` | Preamble: descend from town |
| Modify | `scripts/play.sh` | `ARPG_WORLD_ID=dungeon_levels`, unset `ARPG_SESSION_ID` |
| Modify | `client/scripts/main.gd` | Interactive default world + town wall rendering |
| Modify | `client/tests/test_golden.gd` | Cross-language town/dungeon stair fixture checks |
| Modify | `PROGRESS.md` | v20 lifecycle when complete |

## Task 1 — Shared contracts and world preset

Files:

- Modify: `shared/rules/worlds.v0.json`
- Modify: `shared/protocol/messages.v1.schema.json`
- Modify: `shared/protocol/session_snapshot.v1.schema.json`
- Modify: `shared/protocol/state_delta.v1.schema.json`
- Modify: `shared/golden/dungeon_stairs.json`

- [x] Step 1.1: Update `dungeon_levels` preset with town player spawn, `stairs_down`, `teleporter`.
- [x] Step 1.2: Remove `maximum: -1` (or equivalent) constraints on teleporter level fields;
      allow integer `0` for town.
- [x] Step 1.3: Extend golden fixture with a `town_round_trip` (or update `descend_then_ascend`)
      case: start town → descend to `-1` → descend to `-2` → ascend to `-1` → ascend to town at
      `{8, 10}`.

```bash
make validate-shared
```

## Task 2 — Server town bootstrap and travel

Files:

- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Introduce `townLevel = 0`; multi-level `NewSimWithWorld` starts at `0`, builds
      town `LevelState` from world preset entities (reuse interactable spawn path), sets
      `discoveredTeleporters[0] = true`.
- [x] Step 2.2: Stop pre-generating level `-1` at session create.
- [x] Step 2.3: Update `handleTransition`:
      - `ascend_intent` from `-1` → town at town `stairs_down` (not generated floor).
      - `already_at_entry` when `current_level >= 0`.
      - `descend_intent` from town → lazy `ensureDungeonLevel(-1)`, arrive at `stairs_up`.
- [x] Step 2.4: Update `handleTeleport` to allow `target_level == 0`; resolve town via static
      level, arrive at town teleporter position.
- [x] Step 2.5: Update `teleporterDiscoveryView` to include level `0` when town exists.
- [x] Step 2.6: Add tests:
      - fresh session starts at `0` with preset entities, no `-1` in `levels` until descend;
      - town → `-1` → town via ascend at down stair coords;
      - discover dungeon teleporter → teleport to town;
      - descend through `-1`, `-2`, `-3` without cap.

```bash
cd server && go test ./internal/game/... -run 'Town|Dungeon|Teleport|Play'
```

## Task 3 — Bot scenarios and play entrypoint

Files:

- Modify: `tools/bot/scenarios/12_dungeon_levels.json`
- Modify: `tools/bot/scenarios/13_teleporter_lab.json`
- Modify: `scripts/play.sh`

- [x] Step 3.1: Prepend `{ "action": "use_stair", "direction": "down" }` (or equivalent) to
      scenarios `12` and `13`; add `assert_current_level` / `visited_levels_contain` for `0` where
      useful.
- [x] Step 3.2: Export `ARPG_WORLD_ID=dungeon_levels` and `ARPG_SESSION_ID=` (empty) in `play.sh`
      before launching Godot.
- [x] Step 3.3: Run bot suite.

```bash
make db-up && make bot
```

## Task 4 — Client defaults and town presentation

Files:

- Modify: `client/scripts/main.gd`

- [x] Step 4.1: When not in bot-client mode and `ARPG_WORLD_ID` is empty, default session request
      to `dungeon_levels`.
- [x] Step 4.2: In `_render_world_walls`, for `dungeon_levels` + `current_level == 0`, skip dungeon
      perimeter; re-render on `level_changed`.
- [x] Step 4.3: Confirm waypoint panel lists town (level `0`) as discovered from attach; teleporter
      click on town opens panel (no server discovery needed).

```bash
make client-unit
make client-smoke   # if server available
```

## Task 5 — Lifecycle docs and CI

Files:

- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark v20 complete in lifecycle table; document town entry and play-loop proof.
- [x] Step 5.2: Note deferred gaps (character persistence, safe zone, production town art).

```bash
make ci
```

Manual check:

```bash
make play
# Town start → descend → discover teleporter → return to town → close window (fresh run next time)
```

## Final verification

- [x] `make validate-shared`
- [x] `make test-go`
- [x] `make client-unit`
- [x] `make bot`
- [x] `make ci`
- [ ] Manual `make play` town loop

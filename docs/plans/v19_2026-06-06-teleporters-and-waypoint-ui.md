# v19 Plan — Teleporters and waypoint UI

Status: Complete; `make ci` green on 2026-06-06

## 1. Goal

Add deterministic per-level dungeon teleporters, session-scoped discovery, a server-owned
`teleport_intent`, protocol bot coverage, and a small Godot waypoint panel.

## 2. Baseline and Shortcut Decision

- v18 introduced multi-level dungeons, generated stairs, level-scoped deltas, and a level HUD.
- v19 reuses v18 transition delta ordering instead of adding a new client loading model.
- Discovery stays session-scoped because character-scoped persistence is a larger ADR-0008 backlog
  item.

Godot shortcut adoption checklist:

- **Decision:** reject plugin/addon adoption for v19.
- **Reason:** the client surface is a compact list panel, scroll container, buttons, and one
  placeholder teleporter mesh. Inventory UI plugins are irrelevant and would add authority-risking
  state abstractions.
- **Borrow:** reuse existing in-repo `main.gd` HUD/panel style and interactable click routing.

## 3. File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Teleporter placement config |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate teleporter config |
| Modify | `shared/rules/interactables.v0.json` | Add `teleporter` ready interactable |
| Modify | `shared/rules/interactables.v0.schema.json` | Allow `waypoint` transition |
| Modify | `shared/protocol/messages.v1.schema.json` | Add `teleport_intent` |
| Modify | `shared/protocol/session_snapshot.v1.schema.json` | Add `discovered_teleporters` |
| Modify | `shared/protocol/state_delta.v1.schema.json` | Add discovery update/event fields |
| Modify | `shared/golden/dungeon_stairs.json` | Pin generated teleporter positions |
| Modify | `tools/validate_shared.py` | Validate teleporter placement and schemas |
| Modify | `server/internal/game/dungeon_gen.go` | Generate teleporter entities |
| Modify | `server/internal/game/sim.go` | Discovery state and teleport transition |
| Modify | `server/internal/game/types.go` | Discovery protocol view |
| Modify | `server/internal/inputdecode/inputdecode.go` | Decode `teleport_intent` |
| Modify | `server/internal/game/game_test.go` | Golden and transition tests |
| Modify | `tools/bot/run.py` | Teleporter step/assertions |
| Add | `tools/bot/scenarios/13_teleporter_lab.json` | Bot acceptance scenario |
| Modify | `client/scripts/main.gd` | Teleporter rendering and panel |
| Modify | `client/tests/test_golden.gd` | Golden fixture checks |
| Modify | `docs/PROGRESS.md` | v18/v19 lifecycle update |

## 4. Tasks

### Task 1 — Specs and shared contracts

- [x] Add teleporter placement rules and schema validation.
- [x] Add the `teleporter` interactable with `transition: "waypoint"`.
- [x] Add `teleport_intent`, discovery snapshot state, discovery delta change, and discovery event
  schema support.
- [x] Extend dungeon golden fixture with teleporter positions.

Focused check:

```bash
make validate-shared
```

### Task 2 — Server generation and discovery

- [x] Generate one teleporter per dungeon level using the level-local RNG stream.
- [x] Track `discoveredTeleporters map[int]bool` on `Sim`.
- [x] Include generated/visited levels in snapshots as enabled/disabled teleporter entries.
- [x] Treat `action_intent` on a teleporter as discovery when in range.
- [x] Ack already discovered teleporter actions without changing state.

Focused check:

```bash
cd server && go test ./internal/game/... -run 'Dungeon|Teleport|Rules'
```

### Task 3 — Teleport transition intent

- [x] Add `TeleportIntent{TargetLevel int}` and decode `teleport_intent`.
- [x] Reject non-dungeon worlds, dead players, missing current teleporter reach, invalid levels, and
  undiscovered target levels.
- [x] Move the player to the destination level's teleporter and emit v18-style transition deltas.
- [x] Ensure replay reconstructs discovery and teleport travel from persisted inputs.

Focused check:

```bash
cd server && go test ./internal/game/... -run 'Teleport|Replay'
```

### Task 4 — Bot coverage

- [x] Add bot helpers to discover/open teleporter, assert discovery list, and teleport by level.
- [x] Add `13_teleporter_lab.json`: discover -1, descend, see -2 disabled, discover -2, teleport
  back to -1.
- [x] Verify reconnect state and replay include discovery state.

Focused check:

```bash
make bot
```

### Task 5 — Godot presentation

- [x] Render a simple teleporter placeholder mesh.
- [x] First click sends `action_intent`; already discovered current teleporter opens panel.
- [x] Build left-side panel with level rows, disabled undiscovered rows, and scroll after nine rows.
- [x] Send `teleport_intent` from enabled row clicks.
- [x] Keep panel state in sync after snapshots and deltas.

Focused check:

```bash
make client-unit
make client-smoke
```

### Task 6 — Docs and final verification

- [x] Update `docs/PROGRESS.md` to mark v18 complete and v19 implemented when gates pass.
- [x] Run final CI when DB-dependent infrastructure is available.

Final gate:

```bash
make ci
```

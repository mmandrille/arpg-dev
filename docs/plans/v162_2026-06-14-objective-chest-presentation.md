# v162 Plan — Objective chest presentation

Status: Complete
Goal: Objective chests are visibly distinct in the client while server authority remains unchanged.
Architecture: The server adds optional `elite_objective` metadata to objective chest entity views.
Shared v8 schemas allow the optional boolean, and the Godot client renders a display-only marker
from that metadata. Interaction, lock checks, loot, and objective completion remain server-owned.
Tech stack: Go entity views, shared JSON schemas, Godot client, client bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v161 `full-elite-clear-objective`. Godot plugin/adoption checklist: reject external
plugins/assets; this is a small in-repo presentation marker using the existing primitive chest
model and bot debug patterns.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/types.go` | Optional `elite_objective` entity view field |
| Modify | `server/internal/game/sim.go` | Populate objective metadata for runtime objective chest IDs |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Allow optional snapshot field |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Allow optional delta field |
| Modify | `client/scripts/main.gd` | Objective chest marker, record metadata, bot debug |
| Modify | `client/scripts/bot_scenario_runner.gd` | Presentation assertion support |
| Modify | `client/tests/test_item_visuals.gd` | Unit proof for objective marker and open state |
| Create | `tools/bot/scenarios/client/41_objective_chest_presentation.json` | Client bot proof |
| Create | `docs/as-built/v162_objective-chest-presentation.md` | As-built summary |
| Modify | `docs/specs/v162_spec-objective-chest-presentation.md` | Mark complete at closeout |
| Modify | `PROGRESS.md` | Lifecycle closeout and next slice |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `client/scripts/main.gd`
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Keep the `main.gd` change minimal and offset any line growth if needed; defer larger
  presentation extraction to a dedicated client paydown slice.

Verification:

```bash
make maintainability
```

## Task 1 — Optional server metadata

Files:
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/sim.go`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Modify: `shared/protocol/state_delta.v8.schema.json`

- [x] Add optional `elite_objective` to entity views and schemas.
- [x] Set it only for runtime objective chest IDs.
- [x] Keep ordinary treasure chest views unchanged.

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestEliteObjectiveChestRequiresLeaderKill|TestPopulateDungeonLevelTracksEliteObjectiveChestIDs' -count=1
```

## Task 2 — Client marker and unit coverage

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_item_visuals.gd`

- [x] Store `elite_objective` on interactable records.
- [x] Add a distinct objective marker/ring to objective treasure chests.
- [x] Keep open-lid and inner-glow behavior unchanged.
- [x] Add unit coverage for marker presence and open state.

```bash
make client-unit
```

## Task 3 — Client bot presentation proof

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Create: `tools/bot/scenarios/client/41_objective_chest_presentation.json`

- [x] Expose objective marker state through existing presentation debug rows.
- [x] Add a client bot scenario that descends to the pinned objective floor and asserts the marker.

```bash
make bot-client scenario=41_objective_chest_presentation.json
```

## Task 4 — Lifecycle docs and CI

Files:
- Create: `docs/as-built/v162_objective-chest-presentation.md`
- Modify: `docs/plans/v162_2026-06-14-objective-chest-presentation.md`
- Modify: `docs/specs/v162_spec-objective-chest-presentation.md`
- Modify: `PROGRESS.md`

- [x] Mark completed plan tasks.
- [x] Update spec status/as-built/progress.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestEliteObjectiveChestRequiresLeaderKill|TestPopulateDungeonLevelTracksEliteObjectiveChestIDs' -count=1`
- [x] `make client-unit`
- [x] `make bot scenario=68_dungeon_elite_side_objective.json`
- [x] `make bot-client scenario=41_objective_chest_presentation.json`
- [x] `make ci`

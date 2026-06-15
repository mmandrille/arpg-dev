# v173 Plan — Quest Floor Map Marker

Status: Ready for implementation
Goal: Mark random quest reward chests in protocol and render a distinct client marker on the reward floor.
Architecture: The server remains authoritative by carrying generated `questReward` chest metadata into `LevelState` and projecting it through optional entity-view fields. The client treats `quest_reward` as display-only metadata and delegates marker creation to `chest_presentation.gd`. Bot and debug filters assert the marker without changing chest interaction rules.
Tech stack: Go sim, shared JSON schemas, Godot client scripts/tests, client bot scenario JSON, lifecycle docs.

## Baseline and Shortcut Decision

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Allow optional `quest_reward` on entity views |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Allow optional `quest_reward` on delta entity views |
| Modify | `server/internal/game/types.go` | Add `QuestReward` entity view field |
| Modify | `server/internal/game/level.go` | Track quest reward chest ids per level |
| Modify | `server/internal/game/dungeon_population.go` | Preserve generated quest reward metadata at runtime |
| Modify | `server/internal/game/sim.go` | Emit `QuestReward` in entity views |
| Modify | `server/internal/game/random_quest_floors_test.go` | Prove quest reward chest entity view metadata |
| Modify | `client/scripts/chest_presentation.gd` | Render/check quest marker |
| Modify | `client/scripts/main.gd` | Store flag, pass it to chest presentation, expose bot debug |
| Modify | `client/scripts/bot_scenario_runner.gd` | Filter on quest marker metadata |
| Modify | `client/tests/test_item_visuals.gd` | Unit coverage for quest marker presentation |
| Add | `tools/bot/scenarios/client/42_quest_reward_chest_presentation.json` | Client bot proof |
| Add | `docs/as-built/v173_quest-floor-map-marker.md` | As-built summary |
| Modify | `PROGRESS.md` | Lifecycle update |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_scenario_runner.gd`
- [x] `client/tests/test_item_visuals.gd`
- [x] `server/internal/game/game_test.go`
- [x] `tools/bot/run.py`
- [x] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: `main.gd` already delegates chest marker construction to `chest_presentation.gd`; this slice kept new behavior there and only threaded one display flag through existing entity plumbing.

Verification:
```bash
make maintainability
```

## Task 1 — Protocol and Server Metadata

Files:
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Modify: `shared/protocol/state_delta.v8.schema.json`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/level.go`
- Modify: `server/internal/game/dungeon_population.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/random_quest_floors_test.go`

- [x] Step 1.1: Add optional `quest_reward` to v8 entity schemas and `QuestReward` to `EntityView`.
- [x] Step 1.2: Track quest reward chest ids when generated dungeon chests are populated into a level.
- [x] Step 1.3: Emit `QuestReward` from `entityView` for matching interactable chest ids.
- [x] Step 1.4: Add focused Go coverage for generated quest reward chest entity views.
```bash
make validate-shared
cd server && go test ./internal/game -run 'TestRandomQuest' -count=1
```

## Task 2 — Client Presentation

Files:
- Modify: `client/scripts/chest_presentation.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_item_visuals.gd`

- [x] Step 2.1: Extend chest presentation helpers with `QuestRewardMarker` creation, visibility, and marker-query functions.
- [x] Step 2.2: Thread `quest_reward` through entity parsing, chest creation, opened-state marker sync, and bot/debug entity snapshots.
- [x] Step 2.3: Add client unit coverage for marker creation and opened-state persistence.
```bash
make client-unit
```

## Task 3 — Bot Scenario

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Add: `tools/bot/scenarios/client/42_quest_reward_chest_presentation.json`

- [x] Step 3.1: Add client bot filters for `quest_reward` and `has_quest_marker`.
- [x] Step 3.2: Create a pinned client scenario that descends to `v155_bot_quest_0015` level `-1` and asserts the reward marker.
```bash
make bot scenario=65_random_quest_reward_floor.json
make bot-client scenario=42_quest_reward_chest_presentation.json
```

## Task 4 — Lifecycle Docs and CI

Files:
- Add: `docs/as-built/v173_quest-floor-map-marker.md`
- Modify: `docs/plans/v173_2026-06-14-quest-floor-map-marker.md`
- Modify: `docs/specs/v173_spec-quest-floor-map-marker.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark plan tasks complete and write as-built notes.
- [x] Step 4.2: Update `PROGRESS.md` lifecycle and next-slice pointer.
```bash
make ci
```

## Final Verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestRandomQuest' -count=1`
- [x] `make client-unit`
- [x] `make bot scenario=65_random_quest_reward_floor.json`
- [x] `make bot-client scenario=42_quest_reward_chest_presentation.json`
- [x] `make ci`

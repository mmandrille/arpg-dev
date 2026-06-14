# v162 Spec — Objective chest presentation

Date: 2026-06-14
Status: Complete
Codename: objective-chest-presentation

## Purpose

Make generated elite-objective chests visually distinct in the Godot client so the v158-v161
server-owned side objective is readable before and after completion. The server remains
authoritative; the client only displays optional objective metadata carried on interactable entity
views.

## Non-goals

- No quest journal, minimap pin, objective tracker, NPC turn-in, or durable quest state.
- No new chest loot, monster tuning, objective completion rule, or interaction behavior.
- No external Godot plugin or asset dependency.

## Acceptance criteria

- Snapshot and delta entity views can mark objective chests with optional `elite_objective: true`.
- The client stores that metadata for interactable records without using it for gameplay decisions.
- Objective chests render with a distinct objective crest/ring while ordinary treasure chests keep
  their existing model.
- Opening an objective chest keeps the existing open-lid/glow behavior.
- Client bot presentation debug can assert that an objective chest marker is present.
- Existing protocol bot proof `68_dungeon_elite_side_objective` remains green.

## Scope and likely files

- `server/internal/game/types.go` / `sim.go`: optional entity-view metadata.
- `shared/protocol/session_snapshot.v8.schema.json` and `state_delta.v8.schema.json`: optional
  boolean field.
- `client/scripts/main.gd`: display-only objective chest marker and bot debug state.
- `client/scripts/bot_scenario_runner.gd` and a client scenario: presentation assertion.
- `client/tests/test_item_visuals.gd`: unit coverage for objective chest marker/open state.
- `docs/as-built/v162_objective-chest-presentation.md` and `PROGRESS.md`: lifecycle closeout.

## Test and bot proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestEliteObjectiveChestRequiresLeaderKill|TestPopulateDungeonLevelTracksEliteObjectiveChestIDs' -count=1`
- `make client-unit`
- `make bot scenario=68_dungeon_elite_side_objective.json`
- `make bot-client scenario=41_objective_chest_presentation.json`
- `make ci`

Visual verification command:

```bash
make bot-visual scenario=68_dungeon_elite_side_objective.json
```

## Open questions and risks

- No blocking product questions. The marker should be modest and readable using existing primitive
  mesh patterns.
- Protocol risk is limited to an optional field on existing v8 entity views.

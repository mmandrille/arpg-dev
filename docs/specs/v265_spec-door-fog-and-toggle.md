# v265 Spec - Door Fog And Toggle

Status: Complete
Date: 2026-06-18
Codename: door-fog-and-toggle

## Purpose

Make dungeon doors behave consistently across visibility, navigation, and interaction. A closed door
must block client fog-of-war line-of-sight just like it already blocks server movement and server fog
visibility. Door gaps and the visible door panel should be wide enough to navigate comfortably. An
open door should accept another click to close again.

## Non-goals

- No new door art pipeline, imported assets, Godot addon, or external plugin.
- No durable explored-map memory or reconnect/resume map memory.
- No change to chest one-shot loot semantics.
- No new protocol version; existing `interactable_state_changed.state` events are sufficient.

## Acceptance Criteria

- Closed `wooden_door` interactables are included in the client fog overlay as dynamic occluders, so
  the player cannot visually see through a closed door.
- Opening a door removes that dynamic fog occluder; closing it adds the occluder back.
- Door gap width and closed barrier width are increased through shared rules, not hardcoded server
  constants.
- The client door mesh and pick collider read the same shared barrier width so the visible/clickable
  door matches gameplay tuning.
- Clicking an open wooden door closes it again and emits an authoritative state change.
- Clicking a closed door from one side of a wall should choose an approach point on that same side
  when one is reachable, instead of walking around the wall to interact from behind.
- Treasure chests and other non-barrier interactables keep their existing one-shot/open behavior.
- A focused client bot scenario proves closed -> open -> closed fog occluder behavior.

## Scope and Likely Files

- Shared rules:
  - `shared/rules/interactables.v0.json`
  - `shared/rules/dungeon_generation.v0.json`
- Server:
  - `server/internal/game/interactables.go`
  - `server/internal/game/sim.go`
  - `server/internal/game/interactables_test.go`
- Client:
  - `client/scripts/interactable_rules_loader.gd`
  - `client/scripts/town_node_factory.gd`
  - `client/scripts/main.gd`
  - `client/scripts/bot_assertion_handlers.gd`
  - `client/scripts/bot_scenario_runner.gd`
- Bot:
  - `tools/bot/scenarios/client/73_door_fog_toggle.json`
- Docs:
  - `docs/plans/v265_2026-06-18-door-fog-and-toggle.md`
  - `docs/as-built/v265_door-fog-and-toggle.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets/plugins. Borrow the existing primitive door
presentation and shared-rule loading pattern.

## Test and Bot Proof

```bash
make validate-shared
go test ./server/internal/game -run 'TestDoorLabClosedDoorPreventsPassageUntilActivated|TestOpenDoorCanBeClosedAgain|TestGeneratedDungeonDoorGeneration|TestGeneratedDungeonDoorsPopulateAsClosedInteractables|TestFogOfWarDeltasRevealMonstersWhenClosedDoorOpens'
make client-unit
HEADLESS=1 make bot-visual scenario=73_door_fog_toggle
make maintainability
```

Manual visual proof, if desired:

```bash
make bot-visual scenario=73_door_fog_toggle
make play
```

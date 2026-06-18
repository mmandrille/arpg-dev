# v265 Plan - Door Fog And Toggle

Status: Complete
Goal: Closed doors should block client fog, doors should be wider, and open doors should click closed.
Architecture: Keep door tuning in shared rules. Let the server toggle barrier interactables from
open back to closed with `interactable_state_changed`. Let the client load interactable barrier
rules and derive both door presentation and dynamic fog occluders from those rules.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/interactables.v0.json` | Increase closed door barrier width |
| Modify | `shared/rules/dungeon_generation.v0.json` | Increase generated door gap width |
| Modify | `server/internal/game/interactables.go` | Toggle open barrier interactables closed |
| Add | `server/internal/game/approach.go` | Select closest/same-side action approach goals |
| Modify | `server/internal/game/sim.go` | Treat open barrier interactables as actionable |
| Modify | `server/internal/game/interactables_test.go` | Prove open door can close and chests stay one-shot |
| Add | `client/scripts/interactable_rules_loader.gd` | Load barrier sizes from shared rules |
| Modify | `client/scripts/town_node_factory.gd` | Size door mesh from shared barrier rules |
| Modify | `client/scripts/main.gd` | Sync closed barrier interactables into fog overlay |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Assert extra fog occluder count |
| Modify | `client/scripts/bot_scenario_runner.gd` | Match wait events by `state` |
| Add | `tools/bot/scenarios/client/73_door_fog_toggle.json` | Client proof for closed-open-closed fog behavior |

## Tasks

- [x] Step 1: Update shared door width tuning, server toggle semantics, and same-side door approach.
- [x] Step 2: Load interactable barrier rules in the client and use them for door mesh/pick/fog
  occluders.
- [x] Step 3: Add focused server and client bot coverage.
- [x] Step 4: Update lifecycle docs and run focused verification.

## Verification

```bash
make validate-shared
go test ./internal/game -run 'TestClosedDoorAutoApproachPrefersPlayerSide|TestDoorLabClosedDoorPreventsPassageUntilActivated|TestOpenDoorCanBeClosedAgain|TestGeneratedDungeonDoorGeneration|TestGeneratedDungeonDoorsPopulateAsClosedInteractables|TestFogOfWarDeltasRevealMonstersWhenClosedDoorOpens'
make client-unit
HEADLESS=1 make bot-visual scenario=73_door_fog_toggle
make maintainability
```

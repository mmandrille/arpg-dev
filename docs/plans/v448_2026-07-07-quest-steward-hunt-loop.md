# v448 Plan — Quest Steward Hunt Loop

Status: Ready for implementation
Goal: Hunt monster → trophy → steward family picker → magic+ depth reward.
Architecture: Server-owned hunt roll, target flag, trophy drop, offer generation, and reward roll; client banner + picker only.

## Baseline and shortcut decision

- Reuse v291 quest giver service, v175 tracker placement, mystery-shop roll patterns, `rollItemTemplate` with min-rarity filter.
- Adopt in-repo UI patterns; no external plugins.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/rules/quest_steward.v0.json` | Hunt + reward tuning |
| Create | `shared/rules/quest_steward.v0.schema.json` | Schema |
| Modify | `shared/rules/items.v0.json` | Trophy quest items |
| Modify | `shared/protocol/messages.v8.schema.json` | `quest_steward_pick_intent` |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Hunt events + entity flag |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Entity flag |
| Create | `server/internal/game/dungeon_steward_hunt.go` | Floor roll + target selection |
| Create | `server/internal/game/quest_steward.go` | Offers, pick, trophy drop |
| Create | `server/internal/game/quest_steward_rewards.go` | Family rolls + min rarity |
| Modify | `server/internal/game/quest_turn_in.go` | Route trophy → offers |
| Modify | `server/internal/game/level.go`, `dungeon_population.go`, `sim.go` | Hunt state + views |
| Modify | `server/internal/game/handlers.go`, `inputdecode` | Pick intent |
| Create | `client/scripts/steward_hunt_banner.gd` | Red entry banner |
| Create | `client/scripts/quest_steward_panel.gd` | Family picker |
| Modify | `client/scripts/quest_elite_objective_state.gd` | Journal + tracker |
| Modify | `tools/bot/run.py` | Pick action + assertions |
| Create | `tools/bot/scenarios/120_steward_hunt_quest.json` | Protocol proof |
| Create | `tools/bot/scenarios/client/98_steward_hunt_quest.json` | Client proof |

## Maintenance ratchet

Hotspot files: `main.gd` (minimal wiring), `sim.go` (small fields only).
Extract hunt logic to new files; touch-to-shrink on wiring.

## Task 1 — Shared contracts

- [x] Add quest steward rules, trophies, protocol fields
```bash
make validate-shared
```

## Task 2 — Server hunt + rewards

- [x] Generation, kill drop, turn-in offers, pick handler
```bash
cd server && go test ./internal/game -run 'StewardHunt|QuestSteward' -count=1
```

## Task 3 — Client presentation

- [x] Banner, panel, journal/tracker sync
```bash
make client-unit
```

## Task 4 — Bot scenarios

- [x] Protocol + client extended scenarios
```bash
make bot scenario=120_steward_hunt_quest
make bot-client scenario=98_steward_hunt_quest HEADLESS=1
```

## Task 5 — Lifecycle

- [x] PROGRESS, lifecycle, as-built
```bash
make maintainability
```

## Final verification

Focused autoloop gate (batch CI deferred to post-loop):

```bash
make validate-shared
cd server && go test ./internal/game -run 'StewardHunt|QuestSteward' -count=1
make bot scenario=120_steward_hunt_quest
make bot-client scenario=98_steward_hunt_quest HEADLESS=1
make maintainability
```

# v16 Plan — Use consumable + hotbar

Status: In progress (2026-06-06)

## File map

| Action | Path | Purpose |
|--------|------|---------|
| Add | `shared/protocol/examples/use_intent.json` | Protocol example |
| Modify | `shared/protocol/messages.v0.schema.json` | `use_intent` |
| Modify | `shared/protocol/state_delta.v0.schema.json` | `player_healed`, `item_used` events |
| Modify | `shared/rules/items.v0.json` | `red_potion.heal` |
| Modify | `shared/rules/items.v0.schema.json` | consumable `heal` field |
| Add | `shared/golden/use_consumable.json` | Heal formula fixture |
| Add | `shared/golden/use_consumable.v0.schema.json` | Golden schema |
| Modify | `shared/rules/worlds.v0.json` | `heal_lab` |
| Modify | `server/internal/game/sim.go` | `handleUse` |
| Modify | `server/internal/game/rules.go` | `Category`, `Heal` on items |
| Modify | `server/internal/inputdecode/inputdecode.go` | decode `use_intent` |
| Modify | `server/internal/realtime/runner.go` | persist inventory intents |
| Modify | `server/internal/game/game_test.go` | use consumable tests |
| Add | `client/scripts/consumable_bar.gd` | Bottom hotbar UI |
| Modify | `client/scripts/main.gd` | Wire bar, keys 1–0, bot API |
| Modify | `client/scripts/bot_controller.gd` | Hotbar bot actions |
| Modify | `client/scripts/bot_scenario_runner.gd` | New step types |
| Add | `tools/bot/scenarios/08_heal_lab.json` | Protocol scenario |
| Add | `tools/bot/scenarios/client/06_use_potion_hotbar.json` | Client bot |
| Modify | `tools/bot/run.py` | `use_inventory_item`, assertions |
| Modify | `tools/validate_shared.py` | consumable heal drift guard |
| Modify | `client/tests/test_golden.gd` | use_consumable golden |
| Modify | `PROGRESS.md` | v16 lifecycle |

## Verification

```bash
make validate-shared
make test-go
make client-unit
make bot
make bot-client
make ci
```

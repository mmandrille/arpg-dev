# v406 Plan — Bishop Force Drop

Status: Complete
Goal: Unified Bishop gameplay-debug loot picker with depth cap and forced item level.
Architecture: Server builds loot catalogs from loaded rules (post main_config drop-rate mutation).
Client wizard sends catalog/source/force intents. Resource shortcut intents removed.
Tech stack: Go sim, shared protocol v8, Godot client, Python/Godot bot.

## Baseline and shortcut decision

- Reuse Bishop service, `spawnLootDrops`, resource loot, wallet badge spawn paths.
- Adopt in-repo Bishop panel patterns; extract `bishop_loot_debug_panel.gd`.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/protocol/messages.v8.schema.json` | New catalog/force intents; remove drop shortcuts |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Catalog + force drop events |
| Create | `server/internal/game/bishop_loot_debug.go` | Catalog builder + force handler |
| Modify | `server/internal/game/bishop_debug.go` | Remove drop handlers |
| Modify | `server/internal/game/sim.go` | Intent types, `goldRollContext.forcedItemLevel` |
| Modify | `server/internal/game/types.go` | Catalog view types on Event |
| Modify | `server/internal/inputdecode/inputdecode.go` | Wire new intents |
| Modify | `server/internal/game/handlers.go` | Handler registry |
| Modify | `server/internal/game/item_rolls.go` | Forced item level in template roll |
| Modify | `server/internal/game/resource_loot_drops.go` | Spawn at explicit level |
| Create | `client/scripts/bishop_loot_debug_panel.gd` | Wizard UI |
| Modify | `client/scripts/bishop_panel.gd` | Single force-loot button |
| Modify | `client/scripts/main.gd` | Intent wiring + events |
| Modify | bot runner / controller | Force-loot bot steps |
| Create | `tools/bot/scenarios/115_bishop_force_drop_lab.json` | Extended proof |
| Modify | blacksmith client scenarios | Use force-loot steps |

## Maintenance ratchet

- Extract `bishop_loot_debug_panel.gd` to avoid growing `bishop_panel.gd` past baseline.

## Task 1 — Protocol + types

- [x] Step 1.1: Add intents and event schemas
```bash
make validate-shared
```

## Task 2 — Server catalog + force drop

- [x] Step 2.1: Implement `bishop_loot_debug.go`
- [x] Step 2.2: Remove old drop handlers; forced item level in rolls
```bash
cd server && go test ./internal/game/... -run 'BishopLootDebug|BishopDebug' -count=1
```

## Task 3 — Client wizard

- [x] Step 3.1: `bishop_loot_debug_panel.gd` + panel integration
```bash
godot --headless --path client --script res://tests/test_bishop_panel.gd
```

## Task 4 — Bot scenarios

- [x] Step 4.1: `bishop_force_drop_lab.json`; migrate blacksmith scenarios
```bash
make bot scenario=bishop_force_drop_lab
```

## Task 5 — Lifecycle docs

- [x] Update PROGRESS, lifecycle, as-built

## Final verification

- [x] `make validate-shared`
- [x] `make maintainability`
- [x] Focused tests above

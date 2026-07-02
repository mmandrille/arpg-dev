# v406 As-Built — Bishop Force Drop

Date: 2026-07-02

## What Shipped

- Replaced four Bishop gameplay-debug resource shortcut intents with a unified force-loot wizard:
  depth (capped at `DeepestDungeonDepth`) → source (`monster`, `chest`, `boss`, `boss_chest`) → roll tree
  (treasure entries, resource pool, wallet badges) → optional item level → spawn near player.
- Server catalog builders in `bishop_loot_debug.go`; force handler reuses `spawnLootDrops`, resource loot,
  and wallet badge spawn paths with `forcedItemLevel` on template rolls.
- Client `bishop_loot_debug_panel.gd` wizard; Bishop panel exposes one **Debug: force loot drop…** action.
- Protocol bot `force_bishop_loot` action; extended scenario `bishop_force_drop_lab`; blacksmith client
  scenarios migrated to `click_bishop_force_loot`.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'BishopLootDebug|BishopDebug' -count=1`
- `godot --headless --path client --script res://tests/test_bishop_panel.gd`
- `make bot scenario=bishop_force_drop_lab`

## Deferred

- Visual polish for multi-step loot wizard UX
- Pack promotion for `bishop_force_drop_lab` (extended-only; blacksmith scenarios cover client path)

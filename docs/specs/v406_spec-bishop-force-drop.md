# v406 Spec: Bishop Force Drop

Status: Draft
Date: 2026-07-02
Codename: `bishop-force-drop`
Baseline: v405 `class-specialist-gear`

## Purpose

Replace Bishop gameplay-debug resource drop shortcuts with one **Force loot drop…** wizard
that walks data-driven dungeon loot tables (monster, chest, boss, boss chest), resource-loot pool
entries, and wallet badges. The player picks dungeon depth (capped at the character's deepest
reached depth), drop source, treasure-class roll branches, optional item level, then spawns loot
authoritatively near the player.

## Non-goals

- Forcing template affix values, rarity, or magic find outcomes (item level only).
- Shop, mystery seller, quest, elite-objective, or unique-chest loot sources.
- Removing level / skill point / stat point Bishop debug shortcuts.
- Production Bishop art or player-facing loot filters.

## Acceptance Criteria

- Gameplay debug only (`ARPG_GAMEPLAY_DEBUG`); server rejects when disabled.
- Bishop panel shows **Force loot drop…** instead of four resource shortcut buttons.
- Old `bishop_debug_drop_*` intents are removed; force-loot intents replace them.
- Depth picker lists loot-band depths `1..min(band_max, DeepestDungeonDepth)`; depths above
  `DeepestDungeonDepth` are unavailable.
- Sources: `monster`, `chest`, `boss`, `boss_chest` resolve correct loot tables from rules.
- Catalog exposes treasure-class attempts (success/no_drop weights) and weighted entries with labels.
- Catalog exposes resource-loot pool items and wallet badge items.
- Force intent spawns loot via existing `spawnLootDrops` / resource / wallet spawn paths.
- Template and leveled resource drops honor selected `item_level` within
  `MaxItemLevelForDepth(depth)`.
- Blacksmith client scenarios that used renew-stone / upgrade-shard shortcuts use force-loot bot steps.
- Focused Go tests and extended bot scenario prove gold and template force drops.

## Scope And Likely Files

- Protocol: `messages.v8.schema.json`, `state_delta.v8.schema.json`
- Server: `bishop_loot_debug.go`, `bishop_debug.go`, `handlers.go`, `inputdecode`, `sim.go`, `resource_loot_drops.go`, `types.go`
- Client: `bishop_panel.gd`, `bishop_loot_debug_panel.gd`, `main.gd`, bot facade/runner
- Bot: new extended scenario; update `blacksmith_renew_item`, `blacksmith_upgrade_risk`
- Docs: plan, as-built, lifecycle

## Test And Bot Proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'BishopLootDebug|BishopDebug' -count=1
godot --headless --path client --script res://tests/test_bishop_panel.gd
make bot scenario=bishop_force_drop_lab
```

## Asset And Plugin Decision

- Adopt: Bishop panel, draggable window, existing debug gating.
- Reject: external UI plugins.

## Open Questions

- None blocking (depth cap = `DeepestDungeonDepth`; item level clamped to depth max).

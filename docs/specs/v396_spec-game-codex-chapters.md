# v396 Spec — Game Codex Chapters

Status: Approved for implementation
Date: 2026-06-30
Codename: game-codex-chapters
Baseline: v395 `game-codex-shell` complete

## Purpose

Extend the v395 codex compiler and index with additional chapters:

- **Item families** — presentation family metadata plus aggregated `class_affinities` by `item_type`
- **Resources & crafting** — Upgrade Shard, Renew Stone, upgrade/renew/merge mechanics, max item level formulas
- **Loot & treasure classes** — TC flow explanation, example table, sample loot profiles

## Non-goals

- Per-template item catalog, monster bestiary, uniques/sets encyclopedia
- Protocol/server changes

## Acceptance criteria

- [ ] `build_codex.py` emits item_families, resources, and loot chapters
- [ ] Dagger family page includes rogue attack-speed affinity from rules (semantic range)
- [ ] Resource pages derive drop chances and depth-tier formulas from `main_config` / `dungeon_generation`
- [ ] Treasure-class concept page explains attempt/success/no-drop flow
- [ ] Client bot scenario extended to assert family/resource page content
- [ ] `make gen-codex`, `make validate-shared`, focused tests green

## Test and bot proof

```bash
make gen-codex
make validate-shared
.venv/bin/pytest tools/test_build_codex.py -q
make client-unit
make bot-client SCENARIO=44_codex_foundation HEADLESS=1
```

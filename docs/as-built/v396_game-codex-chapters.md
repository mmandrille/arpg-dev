# v396 As-built — Game Codex Chapters

## What shipped

- Extended `build_codex.py` output with **item families** (class affinity aggregation), **resources & crafting** (upgrade/renew/merge + max ilvl formulas), and **loot & treasure classes** chapters.
- Full index: 6 chapters, 83 pages committed in `codex_index.v0.json`.
- Client bot scenario navigates to Barbarian, Dagger family, Upgrade Shard, and Treasure Class concept pages.

## Proof

```bash
make gen-codex
make validate-shared
.venv/bin/pytest tools/test_build_codex.py -q
make bot-client SCENARIO=44_codex_foundation HEADLESS=1
```

## Deferred

Per-template item catalog, monster bestiary, uniques/sets encyclopedia, in-game hotkey codex.

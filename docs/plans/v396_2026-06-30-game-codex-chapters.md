# v396 Plan — Game Codex Chapters

Goal: Extend codex index with item families, resources/crafting, and loot chapters.

## Tasks

- [x] Regenerate full `codex_index.v0.json` (6 chapters, 83 pages)
- [x] Add `select_codex_page` bot step; extend `44_codex_foundation.json`
- [x] Extend compiler tests for full chapter set

## Verification

```bash
make gen-codex
make validate-shared
.venv/bin/pytest tools/test_build_codex.py -q
make client-unit
make bot-client SCENARIO=44_codex_foundation HEADLESS=1
```

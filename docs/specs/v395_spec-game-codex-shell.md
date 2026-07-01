# v395 Spec — Game Codex Shell

Status: Approved for implementation
Date: 2026-06-30
Codename: game-codex-shell
Baseline: v394 `weapon-slot-families` complete

## Purpose

Add a main-menu **Codex** that opens a read-only reference book UI. v395 ships the shell,
compiler pipeline, and first chapters compiled from shared rules/assets:

- Game concepts (stat/rarity prose overlays)
- Classes (stats, movement/light, tiered actives/passives)
- Skills (kind, costs, summaries)

## Non-goals

- Item families, resources/crafting, treasure-class chapters (v396)
- Parsing `docs/specs/*.md`
- Protocol/server changes, discovery/unlock state, production book art
- External UI plugins (reject; reuse Control-based panels)

## Acceptance criteria

- [ ] `shared/content/codex_overlays.v0.json` + schema for concept/mechanic prose
- [ ] `tools/content/build_codex.py` compiles concepts/classes/skills into `codex_index.v0.json`
- [ ] `make gen-codex` regenerates the index; `make validate-shared` validates schemas
- [ ] Main menu shows **Codex**; panel works without an active session
- [ ] `CodexLoader` + `CodexPanel` render chapter/page navigation from compiled index
- [ ] Godot unit tests for loader + panel; client bot scenario `44_codex_foundation.json`
- [ ] Focused verification green

## Test and bot proof

```bash
make validate-shared
make gen-codex
.venv/bin/pytest tools/test_build_codex.py -q
make client-unit
make bot-client SCENARIO=44_codex_foundation HEADLESS=1
```

## Asset decision

Reject external UI assets; code-native book layout with existing menu styling.

# v395 Plan — Game Codex Shell

Goal: Main-menu Codex with compiler pipeline and concepts/classes/skills chapters.

## Tasks

- [x] Add `codex_overlays`, schemas, and `build_codex.py` (concepts/classes/skills)
- [x] Generate `codex_index.v0.json`; wire `make gen-codex`
- [x] Add `CodexLoader`, `CodexPanel`, main menu wiring
- [x] Bot steps + scenario `44_codex_foundation.json`
- [x] Godot unit tests + CODEMAP entries

## Verification

```bash
make validate-shared
make gen-codex
.venv/bin/pytest tools/test_build_codex.py -q
make client-unit
make bot-client SCENARIO=44_codex_foundation HEADLESS=1
```

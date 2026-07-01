# v395 As-built — Game Codex Shell

## What shipped

- Added schema-backed `codex_overlays.v0.json` for concept/mechanic prose and `codex_index.v0.json` compiler output.
- Added `tools/content/build_codex.py` + `make gen-codex` to compile concepts, classes, and skills from shared rules/assets.
- Main menu **Codex** opens a read-only book UI (`codex_panel.gd`) without an active session.
- Client bot scenario `44_codex_foundation.json` proves menu → codex → back flow.

## Proof

```bash
make validate-shared
make gen-codex
.venv/bin/pytest tools/test_build_codex.py -q
godot --headless --path client --script res://tests/test_codex_loader.gd
godot --headless --path client --script res://tests/test_codex_panel.gd
```

## Deferred to v396

Item families, resources/crafting mechanics pages, treasure-class chapter content.

# v426 Plan — Showme Visual Proof Loop

Spec: [`docs/specs/v426_spec-showme-visual-proof.md`](../specs/v426_spec-showme-visual-proof.md)

## Tasks

- [x] Move `visual_capture.gd` + `showme_item_icons_capture.gd` to `client/scripts/showme/`
- [x] Fix preload to `res://scripts/showme/showme_item_icons_capture.gd`
- [x] Point `render_focus.py` at `client/scripts/showme/visual_capture.gd`
- [x] Symlink `skills/showme/scripts/*.gd` → client copies
- [x] Update file-size baseline path
- [x] Add `tools/test_showme.py` + `client/tests/test_showme_load.gd`
- [x] Capture gear + classes screenshots; register showme load in client-smoke if needed
- [x] Update as-built + PROGRESS

## Verification

```bash
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin
python3 skills/showme/scripts/render_focus.py --focus classes
.venv/bin/pytest tools/test_showme.py -q
godot --headless --path client --script res://tests/test_showme_load.gd
```

# v426 As Built — Showme Visual Proof Loop

Date: 2026-07-04

## Shipped

- Relocated showme capture scripts to `client/scripts/showme/` with fixed `res://` preloads.
- `skills/showme/scripts/*.gd` symlink to client copies; `render_focus.py` targets client path.
- Added `tools/test_showme.py` and `client/tests/test_showme_load.gd` (client-smoke gate).
- Verified captures: `gear` (paladin), `classes`, `item-icons` → `.artifacts/showme/`.

## Verification

```bash
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin
.venv/bin/pytest tools/test_showme.py -q
godot --headless --path client --script res://tests/test_showme_load.gd
```

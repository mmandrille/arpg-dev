# v426 Spec — Showme Visual Proof Loop

Status: Approved (autoloop)  
Date: 2026-07-04  
Codename: showme-visual-proof

## Purpose

Make `/showme` reliable from the Godot `client/` project: fix broken `res://` preloads, capture gear/class screenshots for agent review, and add a lightweight CI probe so regressions fail fast.

## Non-goals

- New showme focus types beyond fixing existing ones.
- Server/protocol changes.
- Full screenshot CI gate (macOS/GUI-dependent); probe validates paths and script load only.

## Acceptance criteria

1. `python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin` writes a PNG under `.artifacts/showme/`.
2. `item-icons` focus loads without preload errors.
3. Showme GDScript lives under `client/scripts/showme/` with valid `res://` paths.
4. Python unit test asserts showme script paths and preload target exist.

## Scope

| Area | Files |
|------|-------|
| Client | `client/scripts/showme/*.gd` |
| Tools/skills | `skills/showme/scripts/render_focus.py`, symlinks to client |
| Tests | `tools/test_showme.py` |
| Maintainability | `.maintainability/file-size-baseline.tsv` path update |

## Test proof

- `python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin`
- `.venv/bin/pytest tools/test_showme.py -q`
- `godot --headless --path client --script res://tests/test_showme_load.gd` (script load smoke)

## Adopt / borrow / reject

- **Borrow** existing `visual_capture.gd` — relocate under `client/scripts/showme/` (reject duplicating logic in skills tree).

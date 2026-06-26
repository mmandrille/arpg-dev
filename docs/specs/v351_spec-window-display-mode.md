# v351 Spec: Window Display Mode Settings

Status: Complete  
Date: 2026-06-26  
Codename: `window-display-mode`  
Baseline: v350 `ci-full-green`

## Purpose

Add three persisted display-mode options in Settings: **windowed**, **exclusive fullscreen**, and
**borderless windowed fullscreen**. Store the choice in the existing local `user://settings.json`
alongside all other client preferences.

## Non-goals

- No server, protocol, shared contract, or golden changes.
- No new resolution presets or graphics-quality behavior changes.
- No OS-level multi-monitor picker or per-monitor persistence.
- No hotkey cycling for display mode (settings panel only in v351).

## Acceptance criteria

- `ClientSettings` owns `window_mode` with supported values:
  `windowed`, `fullscreen`, `windowed_fullscreen`.
- `load()` / `save()` round-trip `window_mode` in `user://settings.json`.
- `apply()` maps modes to Godot `DisplayServer` APIs:
  - `windowed` → sized/centered window using existing resolution logic.
  - `fullscreen` → `WINDOW_MODE_FULLSCREEN`.
  - `windowed_fullscreen` → `WINDOW_MODE_EXCLUSIVE_FULLSCREEN` (borderless).
- Settings panel exposes a three-button row; selection persists and applies immediately.
- Unknown/missing values normalize to `windowed`.
- Bot debug state exposes `window_mode` for headless assertions.
- Focused GDScript unit tests cover normalize, save/load, panel sync, and apply dispatch.

## Scope and likely files

| Area | Files |
|------|-------|
| Settings core | `client/scripts/client_settings.gd` |
| Settings UI | `client/scripts/settings_panel.gd`, `client/scripts/client_audio_bridge.gd` |
| Wiring | `client/scripts/main.gd` |
| Tests | `client/tests/test_window_display_mode_settings.gd`, `scripts/client_smoke.sh` |
| Docs | as-built, lifecycle, `PROGRESS.md` |

## Test and bot proof

```bash
make client-unit
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_window_display_mode_settings.gd
```

## Asset decision

- Adopt: existing `ClientSettings` JSON persistence and settings panel button-row pattern.
- Borrow: camera mode / graphics quality UI wiring from v329.
- Reject: external config plugins; separate config file (all settings stay in `settings.json`).

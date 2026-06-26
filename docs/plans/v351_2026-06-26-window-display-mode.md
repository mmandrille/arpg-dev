# v351 Plan: Window Display Mode Settings

Date: 2026-06-26  
Spec: [`docs/specs/v351_spec-window-display-mode.md`](../specs/v351_spec-window-display-mode.md)

## Tasks

- [x] 1.1 Add `window_mode` constants, normalize/load/save, `set_window_mode()`, and `apply()` dispatch in `client_settings.gd`
- [x] 1.2 Add settings panel label, button row, signal, and sync helpers
- [x] 1.3 Wire `main.gd` handler, `_sync_settings_panel`, debug state, and `client_audio_bridge.gd`
- [x] 1.4 Add `test_window_display_mode_settings.gd` and register in `scripts/client_smoke.sh`
- [x] 1.5 Update as-built, lifecycle, `PROGRESS.md`

## Verification

```bash
make client-unit
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_window_display_mode_settings.gd
```

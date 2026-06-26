# v351 As-Built — Window Display Mode Settings

Date: 2026-06-26  
Spec: [`docs/specs/v351_spec-window-display-mode.md`](../specs/v351_spec-window-display-mode.md)  
Plan: [`docs/plans/v351_2026-06-26-window-display-mode.md`](../plans/v351_2026-06-26-window-display-mode.md)

## Shipped behavior

- **`ClientSettings.window_mode`**: persisted in `user://settings.json` alongside all existing settings.
  Supported values: `windowed`, `fullscreen`, `windowed_fullscreen`.
- **`apply()`** maps modes to Godot `DisplayServer` window modes; windowed mode keeps existing
  resolution/centering logic.
- **Settings panel** adds a Display Mode button row above resolution presets; selection applies
  immediately and persists.
- **Bot debug state** exposes `window_mode` for headless assertions.

## Boundaries

- No server/protocol/shared changes.
- No hotkey cycling or multi-monitor picker.

## Verification

```bash
make client-unit
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_window_display_mode_settings.gd
```

# v266 Spec - Map Transparency Setting

Status: Complete
Date: 2026-06-18
Codename: map-transparency-setting

## Purpose

Let players tune the discovery map overlay transparency from Settings so the compact and fullscreen
map can be made more or less intrusive during play.

## Non-goals

- No server-side or account-synced settings.
- No new map legend, map filters, routefinding, or click-to-navigate behavior.
- No redesign of the Settings panel beyond adding the map transparency slider.
- No external assets, plugins, or addon adoption; this is existing UI and code-native map drawing.

## Acceptance Criteria

- Settings includes a localized `Map transparency` slider whose higher UI value makes the map more
  transparent.
- The slider persists as a local client setting and reloads from `user://settings.json`.
- The discovery minimap/fullscreen map applies the saved opacity immediately when the setting
  changes.
- Existing default map opacity stays unchanged for players without the new saved setting.
- Bot/debug state exposes the applied map opacity so visual scenarios can assert it.
- Focused unit tests cover clamping, persistence, panel sync, and minimap opacity application.

## Scope and Likely Files

- Client settings and UI:
  - `client/scripts/client_settings.gd`
  - `client/scripts/settings_panel.gd`
  - `client/scripts/client_audio_bridge.gd`
  - `client/scripts/main.gd`
- Discovery map:
  - `client/scripts/discovery_minimap.gd`
- Bot/testing:
  - `client/scripts/bot_step_catalog.gd`
  - `client/scripts/bot_action_step_validator.gd`
  - `client/scripts/bot_controller.gd`
  - `client/scripts/bot_scenario_runner.gd`
  - `tools/bot/scenarios/client/74_map_transparency_setting.json`
  - `client/tests/test_audio_settings.gd`
  - `client/tests/test_discovery_minimap.gd`
- Text:
  - `shared/i18n/en.json`
  - `shared/i18n/es.json`

Asset/plugin decision: reject external assets/plugins. The work uses the existing Settings panel
slider pattern and the existing code-native discovery map drawing.

## Test and Bot Proof

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_audio_settings.gd
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_discovery_minimap.gd
make validate-shared
make client-unit
HEADLESS=1 make bot-visual scenario=74_map_transparency_setting
make maintainability
```

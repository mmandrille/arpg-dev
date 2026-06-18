# v266 Plan - Map Transparency Setting

Status: Complete
Goal: Add a local Settings slider that controls discovery map overlay opacity.
Architecture: Reuse the existing settings persistence and slider UI patterns. Keep the setting
client-local and apply it directly to `DiscoveryMinimap` drawing/style state.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/client_settings.gd` | Persist and clamp `map_opacity` with existing local settings |
| Modify | `client/scripts/settings_panel.gd` | Add localized map transparency slider and signal |
| Modify | `client/scripts/client_audio_bridge.gd` | Pass the saved value through existing settings-panel show/sync helpers |
| Modify | `client/scripts/discovery_minimap.gd` | Replace fixed opacity with a clamped instance value |
| Modify | `client/scripts/main.gd` | Apply setting to the map and expose a bot setter/debug value |
| Modify | `client/scripts/bot_step_catalog.gd` | Register `set_map_opacity` bot action |
| Modify | `client/scripts/bot_action_step_validator.gd` | Validate the bot action value |
| Modify | `client/scripts/bot_controller.gd` | Dispatch the bot action to the main scene |
| Modify | `client/scripts/bot_scenario_runner.gd` | Include readable step details |
| Modify | `client/tests/test_audio_settings.gd` | Cover persistence, clamps, and settings-panel slider sync |
| Modify | `client/tests/test_discovery_minimap.gd` | Cover minimap opacity setter/debug application |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Add localized label |
| Add | `tools/bot/scenarios/client/74_map_transparency_setting.json` | Visual bot proof for applied opacity |

## Tasks

- [x] Step 1: Add local `map_opacity` persistence and Settings panel slider.
- [x] Step 2: Apply opacity to the discovery map panel and map background.
- [x] Step 3: Add focused unit and bot coverage.
- [x] Step 4: Update lifecycle docs and run focused verification.

## Verification

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_audio_settings.gd
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_discovery_minimap.gd
make validate-shared
make client-unit
HEADLESS=1 make bot-visual scenario=74_map_transparency_setting
make maintainability
```

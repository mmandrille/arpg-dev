# v266 As-Built - Map Transparency Setting

Date: 2026-06-18
Spec: [`docs/specs/v266_spec-map-transparency-setting.md`](../specs/v266_spec-map-transparency-setting.md)
Plan: [`docs/plans/v266_2026-06-18-map-transparency-setting.md`](../plans/v266_2026-06-18-map-transparency-setting.md)

## Shipped Behavior

- Settings now includes a localized `Map transparency` slider; moving it right makes the map more
  transparent.
- The setting is persisted locally as `map_opacity` in `user://settings.json`, with the existing
  default opacity preserved for players without a saved value.
- `DiscoveryMinimap` now applies the clamped opacity to both the panel style and map background
  draw pass.
- Opening Settings syncs the slider from the saved value, and slider changes apply immediately to
  the current map overlay.
- Client bot support includes `set_map_opacity`, and minimap debug state reports the applied
  `panel_opacity`.
- `tools/bot/scenarios/client/74_map_transparency_setting.json` proves low and high opacity values
  are reflected by the active discovery map.

## Boundaries

- No account-synced settings, server contract, or protocol change shipped.
- No map legend/filter/routing behavior shipped.
- No Settings panel redesign or external asset/plugin adoption shipped.

## Verification

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_audio_settings.gd
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_discovery_minimap.gd
make validate-shared
make client-unit
HEADLESS=1 make bot-visual scenario=74_map_transparency_setting
make maintainability
```

All focused commands passed on 2026-06-18.

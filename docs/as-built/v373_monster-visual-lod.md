# v373 As-Built — Monster Visual LOD

## What shipped

- `shared/rules/main_config.v0.json` adds `presentation_lod` (`min_live_monsters`, `distance_threshold`).
- `entity_presentation_lod.gd` disables shadow/GI on distant monsters when crowded.
- Wired from `main.gd` `_process` refresh and entity upsert paths.

## Verification

```bash
godot --headless --path client --script res://tests/test_entity_presentation_lod.gd
make bot scenario=crowded_melee_perf_probe
```

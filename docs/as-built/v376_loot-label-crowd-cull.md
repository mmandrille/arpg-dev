# v376 As-Built — Loot Label Crowd Cull

## What shipped

- `shared/rules/main_config.v0.json` adds `loot_labels` crowd-cull keys (`crowd_cull_min_loot`, `max_visible_labels`, `combat_radius`).
- `main.gd` `_refresh_loot_label_visibility()` ranks labels by distance and hides excess loot nameplates in combat crowds.
- Hold-to-reveal (`ALT`) behavior preserved for filtered labels.

## Verification

```bash
make validate-shared
make bot-client SCENARIO=loot_label_filter HEADLESS=1
```

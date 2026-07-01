# v392 As-built — Class Creation Features

## What shipped

- `class_creation_summary.gd` derives per-class feature lines from `character_progression.v0.json` and `skills.v0.json`.
- Character create panel shows a **Class features** summary when selecting a class (stats, move/light, tier actives, passives).
- Client unit + extended bot `20_menu_create_join_flow` assert feature lines.

## Proof

```bash
make client-unit
make bot-client SCENARIO=20_menu_create_join_flow HEADLESS=1
```

# v392 Plan — Class Creation Features

Date: 2026-06-30
Spec: `docs/specs/v392_spec-class-creation-features.md`

## Tasks

- [x] Shared: per-class creation summary metadata in `character_progression.v0.json` (skill id refs by tier)
- [x] Client: `class_creation_summary.gd` derives feature lines from progression + skills catalogs
- [x] Client: `character_select_panel.gd` features panel on class selection; trim hardcoded `CLASS_DEFS` duplicates
- [x] Client: `get_debug_state()` exposes selected class feature lines for headless tests and client bot
- [x] Tests: client unit coverage for class switch + extended bot `20_menu_create_join_flow` feature summary

## Verification

```bash
make validate-shared
make client-unit
make bot-client SCENARIO=20_menu_create_join_flow HEADLESS=1
```

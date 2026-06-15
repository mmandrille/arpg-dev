# v71 Plan — Class Picker and Sprites

Status: Complete
Goal: Add class selection, class sprites, and class row icons to the Godot character picker.
Architecture: Server authority is already complete in v69/v70. v71 is client presentation and
request plumbing only: the panel sends `character_class` during create, renders class icons from a
small in-repo control, and exposes debug/bot state for proof.
Tech stack: Godot GDScript UI/tests, existing REST character API, client bot scenario, docs.

## Baseline and Shortcut Decision

panel affordance and the project already has a code-native `SkillIcon` drawing pattern to borrow.
Borrow pattern from `client/scripts/skill_icon.gd` for lightweight class sprites.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/class_icon.gd` | Draw code-native class sprites |
| Modify | `client/scripts/character_select_panel.gd` | Class picker blocks, tooltips, row icons, debug state |
| Modify | `client/scripts/net_client.gd` | Send optional `character_class` in create request |
| Modify | `client/scripts/main.gd` | Pass selected class through create signal |
| Modify | `client/tests/test_coop_client.gd` | Unit proof for picker selection and row icons |
| Modify | `client/scripts/bot_scenario_runner.gd`, `tools/bot/scenarios/client/*.json` | Bot proof for non-default class create |
| Add | `docs/as-built/v71_class-picker-and-sprites.md` | Close-out summary |
| Modify | `PROGRESS.md`, this plan | Lifecycle close-out |

## Task 1 — Client Picker UI

Files:
- Add: `client/scripts/class_icon.gd`
- Modify: `client/scripts/character_select_panel.gd`

- [x] Step 1.1: Add class icon control with three sprite variants.
- [x] Step 1.2: Add selectable class blocks under the name input, defaulting to barbarian.
- [x] Step 1.3: Add tooltip text with class stats and class skill.
- [x] Step 1.4: Add class icon at the start of each character row.
- [x] Step 1.5: Expose selected class and row class data in debug state.

```bash
make client-unit
```

## Task 2 — Create Request Plumbing

Files:
- Modify: `client/scripts/net_client.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Change create signal handling to pass `character_class`.
- [x] Step 2.2: Include selected `character_class` in `POST /v0/characters`.
- [x] Step 2.3: Preserve default behavior for existing code paths.

```bash
make client-unit
```

## Task 3 — Tests and Client Bot

Files:
- Modify: `client/tests/test_coop_client.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `tools/bot/scenarios/client/20_menu_create_join_flow.json`

- [x] Step 3.1: Add unit assertions for default/changed class selection, tooltip/debug data, and row icon class.
- [x] Step 3.2: Add bot step support to select a class in the character panel if needed.
- [x] Step 3.3: Update a menu create scenario to select Sorcerer before create and assert the flow still starts.

```bash
make client-unit
make client-smoke
```

## Task 4 — Lifecycle Docs and CI

Files:
- Add: `docs/as-built/v71_class-picker-and-sprites.md`
- Modify: `docs/specs/v71_spec-class-picker-and-sprites.md`
- Modify: `docs/plans/v71_2026-06-11-class-picker-and-sprites.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark spec/plan complete and add as-built summary.
- [x] Step 4.2: Update `PROGRESS.md` lifecycle table, summary, and next slice.
- [x] Step 4.3: Run final CI.

```bash
make ci
```

## Final Verification

- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make ci`

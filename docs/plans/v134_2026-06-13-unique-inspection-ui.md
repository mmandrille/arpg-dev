# v134 Plan — Unique Inspection UI

Status: Complete
Goal: Display readable unique-effect descriptions at the bottom of client item tooltips.
Architecture: Keep effect text data-owned in `shared/rules/unique_effects.v0.json`, load it through
the existing Godot rule loader, and add a small formatter used by the current tooltip builders.
Tech stack: Godot GDScript client UI/tests, lifecycle docs.

## Baseline And Shortcut Decision

is a narrow extension of the existing item tooltip panel and shared rule loader, with no new inventory
presentation system, art pipeline, or external UI widget need.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/item_rules_loader.gd` | Load unique effect rule catalog |
| Create | `client/scripts/unique_effect_tooltip.gd` | Resolve effect ids into tooltip lines |
| Modify | `client/scripts/inventory_panel.gd` | Append effect descriptions at tooltip bottom |
| Modify | `client/scripts/stash_panel.gd` | Include effect descriptions in stash/chest tooltips |
| Modify | `client/scripts/market_panel.gd` | Include effect descriptions in market item tooltips |
| Modify | `client/tests/test_shop_panel.gd` | Cover inventory and market effect tooltip text |
| Modify | `client/tests/test_stash_panel.gd` | Cover unique chest effect tooltip text |
| Create | `docs/as-built/v134_unique-inspection-ui.md` | As-built summary |
| Modify | `docs/specs/v134_spec-unique-inspection-ui.md` | Status closeout |
| Modify | `docs/plans/v134_2026-06-13-unique-inspection-ui.md` | Checkbox closeout |
| Modify | `PROGRESS.md` | Lifecycle and next-slice update |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/inventory_panel.gd`
- [x] `client/scripts/stash_panel.gd`
- [x] `client/scripts/market_panel.gd`

Decision:
- [x] Keep hotspot edits to one-line formatter calls where practical.
- [x] Put reusable effect formatting in a new small script rather than expanding each panel.

Verification:
```bash
make maintainability
```

## Task 1 — Load Unique Effect Rules

Files:
- Modify: `client/scripts/item_rules_loader.gd`

- [x] Step 1.1: Add a cached `unique_effects` dictionary.
- [x] Step 1.2: Load `shared/rules/unique_effects.v0.json` during `ensure_loaded()`.
- [x] Step 1.3: Add a small lookup helper for effect ids.

## Task 2 — Shared Tooltip Formatter

Files:
- Create: `client/scripts/unique_effect_tooltip.gd`

- [x] Step 2.1: Resolve item `effect_ids` arrays into display names and summaries.
- [x] Step 2.2: Return text/rich tooltip lines that can be appended by each panel.
- [x] Step 2.3: Gracefully handle missing or malformed effect ids.

## Task 3 — Panel Wiring

Files:
- Modify: `client/scripts/inventory_panel.gd`
- Modify: `client/scripts/stash_panel.gd`
- Modify: `client/scripts/market_panel.gd`

- [x] Step 3.1: Append inventory effect text after stats, requirements, and comparisons in plain
  tooltip output.
- [x] Step 3.2: Append rich effect lines to the shared tooltip panel content.
- [x] Step 3.3: Add the same effect lines to stash, unique chest, and market tooltip builders.

## Task 4 — Client Tests

Files:
- Modify: `client/tests/test_shop_panel.gd`
- Modify: `client/tests/test_stash_panel.gd`

- [x] Step 4.1: Test inventory unique tooltip text includes `Everburning Wound` and its summary at
  the bottom.
- [x] Step 4.2: Test market tooltip lines include the readable effect summary.
- [x] Step 4.3: Test unique chest/stash tooltip lines include the readable effect summary.

```bash
make client-unit
```

## Task 5 — Lifecycle Docs And CI

Files:
- Create: `docs/as-built/v134_unique-inspection-ui.md`
- Modify: `docs/specs/v134_spec-unique-inspection-ui.md`
- Modify: `docs/plans/v134_2026-06-13-unique-inspection-ui.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark the spec and plan complete.
- [x] Step 5.2: Record v134 completion and next slice in `PROGRESS.md`.
- [x] Step 5.3: Add the v134 as-built summary.

```bash
make maintainability
make ci
```

## Final Verification

- [x] `make client-unit`
- [x] `make maintainability`
- [x] `make ci`

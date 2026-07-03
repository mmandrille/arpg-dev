# v419 As-Built — Skill Tree Prerequisite Layout

Date: 2026-07-03
Spec: [`docs/specs/v419_spec-skill-tree-prerequisite-layout.md`](../specs/v419_spec-skill-tree-prerequisite-layout.md)
Plan: [`docs/plans/v419_2026-07-03-skill-tree-prerequisite-layout.md`](../plans/v419_2026-07-03-skill-tree-prerequisite-layout.md)

## What shipped

- `SkillTreeLayout` resolves combat-active columns from prerequisite chains (same column hint or sole sibling), fan spread, and false-stack avoidance for tier-1 roots.
- Passives shift to column 7+ when a class needs five combat columns; survival stays at column 6.
- Skills panel widened (`SKILL_TREE_WIDTH` 778) to fit the expanded ranger/paladin trees.
- Headless proofs: `test_skill_tree_layout.gd` + ranger assertions in `test_skills_panel.gd`.
- `validate_skills.py` checks chain column-hint tier ordering.

## Proof

```bash
make client-unit
make validate-shared
make maintainability
```

Visual: `/showme skills` with ranger selected.

## Boundaries

- Client presentation only; no gameplay, protocol, or server changes.
- `tree.column` remains a layout hint in shared JSON; resolver owns final combat columns.

# v411 As-Built — Skill Tree Graph Layout

Date: 2026-07-02
Spec: [`docs/specs/v411_spec-skill-tree-graph-layout.md`](../specs/v411_spec-skill-tree-graph-layout.md)
Plan: [`docs/plans/v411_2026-07-02-skill-tree-graph-layout.md`](../plans/v411_2026-07-02-skill-tree-graph-layout.md)

## What shipped

- Extracted `SkillTreeLayout` for authoritative `tree.tier` / `tree.column` grid positions from shared rules.
- Extracted `SkillTreeConnectors` to draw orthogonal prerequisite lines from `requirements.skills[]`.
- Removed hardcoded vertical `ColorRect` connector and fixed disabled-slot placeholders.
- Fixed hover tooltips to show rank/body for the hovered skill instead of a stale selection.
- Added `connections` and `tooltip_rank_label` to skills panel debug state for headless proof.

## Proof

```bash
godot --headless --path client --script res://tests/test_skills_panel.gd
# 148 passed, 0 failed
```

`skills_panel.gd` reduced to 1001 lines (under 1010 baseline).

## Boundaries

- Client presentation only; no protocol, server, or shared rule changes.
- Code-native connector drawing; no external assets.
- Visual check: `/showme skills` or `make bot-visual scenario=19_skill_points_and_magic_bolt`.

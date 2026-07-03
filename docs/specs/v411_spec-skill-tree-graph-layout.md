# v411 Spec — Skill Tree Graph Layout

Status: Complete
Date: 2026-07-02
Codename: skill-tree-graph-layout

## Purpose

Refactor the Skills panel so every class skill tree renders as a **real graph**:

- Node positions come from authoritative `tree.tier` and `tree.column` in `shared/rules/skills.v0.json` (no per-row centering or visible-index column reassignment for active skills).
- Every `requirements.skills[]` prerequisite draws an orthogonal connector from parent to child (active→active, passive→passive, mixed when applicable).
- Connectors render behind skill icons; met prerequisites are brighter than unmet.
- Hover tooltips show rank/stats for the **hovered** skill, not a stale selected skill.

## Non-goals

- No new skills, balance changes, or prerequisite rule edits.
- No protocol/schema bump, server sim changes, or golden updates.
- No production art, zoom/pan, draggable nodes, or respec UI.
- No new bot pack scenario (headless `test_skills_panel.gd` owns layout proof).

## Acceptance criteria

- [ ] Every visible skill block position = `origin + (column-1)*spacing.x`, `origin + (tier-1)*spacing.y` from shared `tree` metadata.
- [ ] Rogue Fan of Blades (`tier:2, column:3`) aligns under column 3; Poison Stab (`tier:1, column:1`) at column 1; connector drawn between them.
- [ ] All visible `requirements.skills[]` edges for the current class render as connectors.
- [ ] Hardcoded placeholder `ColorRect` line and fixed disabled-slot placeholders removed.
- [ ] Hover tooltip rank/body matches the hovered skill's progression row.
- [ ] `make client-unit` green; `skills_panel.gd` does not grow beyond maintainability baseline (extract layout/connectors).

## Scope and likely files

- **New** `client/scripts/skill_tree_layout.gd` — grid position math.
- **New** `client/scripts/skill_tree_connectors.gd` — edge list + draw layer.
- **Modify** `client/scripts/skills_panel.gd` — orchestration, tooltip hover fix.
- **Modify** `client/tests/test_skills_panel.gd` — grid assertions + connector debug state.
- **Docs** `docs/as-built/v411_skill-tree-graph-layout.md`, lifecycle row.

## Test and bot proof

- `make client-unit` — `test_skills_panel.gd` asserts per-class grid positions and `connections` debug entries.
- Visual: `/showme skills` or `make bot-visual scenario=19_skill_points_and_magic_bolt` (manual; not CI gate).

## Client asset decision

**Reject** external assets/plugins. Use code-native `Control._draw()` connectors (same pattern family as in-repo UI primitives).

## Open questions and risks

- `skills_panel.gd` is grandfathered at 1010 lines; extraction is required in this slice.
- Paladin compact-reflow unit test must be replaced with authoritative column assertions.

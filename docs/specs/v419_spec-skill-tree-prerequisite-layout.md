# v419 Spec — Skill Tree Prerequisite Layout

Status: Complete
Date: 2026-07-03
Codename: skill-tree-prerequisite-layout

## Purpose

Extend the v411 skill-tree graph so **display columns follow the skill-prerequisite graph**, not
misleading static `tree.column` coordinates. Players should read vertical stacks as true chains
(Piercing Shot → Volley → Rain of Arrows; Snipe → Explosive Shot) and fan-outs as siblings
(Pinning Shot, Snipe branching from Piercing Shot) without a tier-1 root sitting above an unrelated
tier-2 skill in the same column (Pinning Shot under Disengage; Snipe under Companion).

`tree.tier` remains authoritative for rows. `tree.column` becomes a **layout hint** (root order and
fan sibling order). Passives keep a fixed right column; survival skills keep a dedicated branch
column.

## Non-goals

- No balance, prerequisite rule, synergy, or skill catalog changes.
- No protocol/schema bump, server sim, or golden updates.
- No zoom/pan, draggable nodes, or production art.
- No new CI pack bot scenario (headless layout tests own the contract).

## Acceptance criteria

- [ ] `SkillTreeLayout.block_position` uses a per-class resolved column for combat actives.
- [ ] Ranger: Pinning Shot column ≠ Disengage column; connector Piercing Shot → Pinning Shot unchanged.
- [ ] Ranger: Snipe column ≠ Companion column; Explosive Shot shares Snipe’s column.
- [ ] Ranger: Volley and Rain of Arrows share Piercing Shot’s column (chain).
- [ ] Rogue: Fan of Blades column aligns with Poison Stab chain/fan rules (not Dash column).
- [ ] All classes: passive blocks remain in the fixed right column (x unchanged from pre-slice baseline).
- [ ] Survival skills remain in the branch column area.
- [ ] No false vertical stack: tier-1 root at column C must not share C with a non-descendant at a lower tier.
- [ ] `make client-unit` green; new `test_skill_tree_layout.gd` covers ranger + rogue cases.
- [ ] `validate_skills.py` adds layout-coherence checks for chain children sharing a primary prereq column hint.

## Scope and likely files

- **Modify** `client/scripts/skill_tree_layout.gd` — prerequisite-aware column resolver.
- **Create** `client/tests/test_skill_tree_layout.gd` — focused layout assertions.
- **Modify** `client/tests/test_skills_panel.gd` — ranger alignment checks; use resolved layout helpers.
- **Modify** `client/scripts/skills_panel.gd` — widen active tree width only if resolver needs a fifth combat column.
- **Modify** `tools/validate_skills.py` — chain column-hint coherence checks.
- **Modify** `scripts/client_smoke.sh` — register new unit test.
- **Docs** lifecycle + as-built.

## Test and bot proof

```bash
make client-unit
make validate-shared
```

Visual: `/showme skills` with ranger selected.

## Client asset decision

**Reject** external assets/plugins. Reuse v411 code-native `Control._draw()` connectors.

## Open questions and risks

- Five combat columns may require a modest panel width bump; prefer widening over compressing passive column.

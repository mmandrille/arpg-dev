# v411 Plan — Skill Tree Graph Layout

Status: Ready for implementation
Goal: Render class skill trees as a tier/column grid with prerequisite connectors from shared rules.
Architecture: Extract `SkillTreeLayout` (pure grid math) and `SkillTreeConnectors` (draw layer). `SkillsPanel` owns interaction/tooltips and delegates layout/edges. Connector style: orthogonal L-shape (parent bottom → child top).
Tech stack: Godot 4 GDScript client only; shared rules already own `tree` + `requirements.skills`.

## Baseline and shortcut decision

Builds on v309 passive column (`tree.column` for passives) and v59 data-driven skill catalog. **Reject** external assets; code-native draw.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/skill_tree_layout.gd` | Grid positions + prerequisite edge list |
| Create | `client/scripts/skill_tree_connectors.gd` | Connector draw layer |
| Modify | `client/scripts/skills_panel.gd` | Use layout/connectors; fix hover tooltip |
| Modify | `client/tests/test_skills_panel.gd` | Grid + connection assertions |
| Create | `docs/as-built/v411_skill-tree-graph-layout.md` | Proof summary |

## Maintenance ratchet

Hotspot: `client/scripts/skills_panel.gd` (baseline 1010, currently over).

- [x] Extract `skill_tree_layout.gd` and `skill_tree_connectors.gd`
- [x] Shrink `skills_panel.gd` below 1010 lines

```bash
make maintainability
```

## Task 1 — Layout module

Files:
- Create: `client/scripts/skill_tree_layout.gd`

- [x] Step 1.1: `block_position(skill_id, origin, spacing, block_size)` from `tree.tier/column`
- [x] Step 1.2: `prerequisite_edges(visible_ids, skill_progression)` → `[{from, to, met}]`

## Task 2 — Connector layer

Files:
- Create: `client/scripts/skill_tree_connectors.gd`

- [x] Step 2.1: Control with `set_edges` + `_draw()` orthogonal paths
- [x] Step 2.2: Met/unmet color modulation

## Task 3 — Skills panel integration

Files:
- Modify: `client/scripts/skills_panel.gd`

- [x] Step 3.1: Replace `_skill_block_position` centering with `SkillTreeLayout`
- [x] Step 3.2: Remove hardcoded line + disabled slots; add connector layer behind nodes
- [x] Step 3.3: Hover tooltip uses hovered `skill_id` without stale rank
- [x] Step 3.4: Expose `connections` in `get_debug_state()`

```bash
make maintainability
```

## Task 4 — Unit tests

Files:
- Modify: `client/tests/test_skills_panel.gd`

- [x] Step 4.1: Rogue/paladin/sorcerer grid position assertions from rules
- [x] Step 4.2: Connector count + fan_of_blades→poison_stab edge
- [x] Step 4.3: Hover tooltip rank matches hovered skill

```bash
make client-unit
```

## Task 5 — Lifecycle docs

- [x] `docs/as-built/v411_skill-tree-graph-layout.md`
- [x] `PROGRESS.md` + slice lifecycle row

## Final verification

- [x] `make maintainability`
- [x] `make client-unit`

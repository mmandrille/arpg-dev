# v419 Plan — Skill Tree Prerequisite Layout

Status: Complete
Goal: Resolve skill-tree display columns from prerequisite chains/fans so grid position matches connectors.
Architecture: `SkillTreeLayout` caches per-class resolved `{tier, column}` for combat actives using chain inheritance (same column hint as parent), fan spread (siblings sorted by `tree.column` hint), and false-stack avoidance for tier-1 roots. Passives and survival keep fixed display columns from JSON hints.
Tech stack: GDScript client, shared `skills.v0.json` hints, Python `validate_skills.py`.

## Baseline and shortcut decision

Builds on v411 graph connectors and v418 synergies. **Reject** external assets; code-native layout only.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/skill_tree_layout.gd` | Column resolver + false-stack guard |
| Create | `client/tests/test_skill_tree_layout.gd` | Ranger/rogue/all-class smoke |
| Modify | `client/tests/test_skills_panel.gd` | Ranger column assertions |
| Modify | `client/scripts/skills_panel.gd` | Panel width if needed |
| Modify | `tools/validate_skills.py` | Chain column-hint checks |
| Modify | `scripts/client_smoke.sh` | Register layout test |
| Create | `docs/as-built/v419_skill-tree-prerequisite-layout.md` | Proof summary |

## Maintenance ratchet

Hotspot files touched:
- [ ] `client/scripts/skill_tree_layout.gd` (new logic, stays <600)
- [ ] `client/scripts/skills_panel.gd` (width constant only if needed)

Verification:
```bash
make maintainability
```

## Task 1 — Layout resolver

Files:
- Modify: `client/scripts/skill_tree_layout.gd`

- [x] Add per-class cache, passive/survival fixed columns, chain/fan/false-stack algorithm
- [x] `block_position` / `block_center` use resolved tree
```bash
# run after Task 2
```

## Task 2 — Unit tests

Files:
- Create: `client/tests/test_skill_tree_layout.gd`
- Modify: `client/tests/test_skills_panel.gd`
- Modify: `scripts/client_smoke.sh`

- [x] Ranger + rogue acceptance cases from spec
- [x] All-class false-stack smoke
```bash
make client-unit
```

## Task 3 — Shared validation

Files:
- Modify: `tools/validate_skills.py`

- [x] Chain column-hint coherence for skills with a single skill prereq
```bash
make validate-shared
```

## Task 4 — Panel width (if needed)

Files:
- Modify: `client/scripts/skills_panel.gd`

- [x] Bump `SKILL_ACTIVE_TREE_WIDTH` / panel min width only when fifth combat column is required
```bash
make client-unit
```

## Task 5 — Lifecycle docs

- [x] `docs/as-built/v419_skill-tree-prerequisite-layout.md`
- [x] `PROGRESS.md` + lifecycle row on `/finish`

## Bot scenarios

Deferred — client unit tests own layout contract (spec non-goal).

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `make client-unit`

# v401 Plan — Class Skill Decuple Core

Spec: [`docs/specs/v401_spec-class-skill-decuple-core.md`](../specs/v401_spec-class-skill-decuple-core.md)

## Task 1 — Shared skill catalog

- [x] Update `shared/rules/skills.v0.json`: remove `split_arrow`, `arcane_orb`; rename `ligthing` → `lightning`; add new skills; tier-4 passives
- [x] Update `shared/assets/skill_presentations.v0.json` and i18n keys
- [x] Update codex entries

**Verify:** `make validate-shared`

## Task 2 — Server migration and tests

- [x] Add `ligthing` → `lightning` rank alias in `normalizeSkillRanks`
- [x] Update Go tests referencing removed/renamed skills
- [x] Add focused tests for `skullcrusher`, `consecrated_smite`, tier-4 passives

**Verify:** `cd server && go test ./internal/game/... -run 'Skill|Passive|Tier|Lightning' -count=1`

## Task 3 — Client and tooling

- [ ] Update GDScript tests and `projectile_visuals.gd` for `lightning`
- [ ] Update `tools/validate_skills.py` passive chains + decuple counts
- [ ] Update/remove bot scenarios (`64_ranger_split_arrow`, `ligthing` refs)

**Verify:** `make client-unit`; `.venv/bin/pytest tools/bot/test_skill_visual.py -q`

## Task 4 — Bot proof

- [ ] Extend barbarian/paladin foundation scenarios for new actives

**Verify:** `make bot scenario=51_barbarian_class_foundation.json`; paladin scenario if present

## Task 5 — Docs

- [ ] `docs/as-built/v401_class-skill-decuple-core.md`

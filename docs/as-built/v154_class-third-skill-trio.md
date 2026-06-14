# v154 As-Built - Class Third Skill Trio

Spec: [`docs/specs/v154_spec-class-third-skill-trio.md`](../specs/v154_spec-class-third-skill-trio.md)
Plan: [`docs/plans/v154_2026-06-14-class-third-skill-trio.md`](../plans/v154_2026-06-14-class-third-skill-trio.md)

## What shipped

- Added Barbarian `earthbreaker`, Rogue `shadow_flurry`, and Ranger `split_arrow` to the shared
  skill catalog.
- Kept all existing skill definitions unchanged.
- Added skill presentation metadata and English/Spanish text for the new skills.
- Extended shared validation so the new class skills keep their expected class ownership and
  prerequisite chains.
- Added focused Go tests for catalog shape and prerequisite spendability.
- Extended Godot skill loader and skill panel tests so the new higher-row skills are visible and
  gated correctly.
- Added protocol bot proofs:
  - `make bot scenario=62_barbarian_earthbreaker.json`
  - `make bot scenario=63_rogue_shadow_flurry.json`
  - `make bot scenario=64_ranger_split_arrow.json`

## Key decisions

- Reused existing skill capability types (`cone_attack` and `projectile_attack`) to keep this slice
  focused on player-visible class expansion without a new mechanics framework.
- Used focused per-class bot scenarios because the current protocol bot scenario shape creates one
  character per scenario.
- Rejected external Godot skill-tree/VFX plugins; the existing shared-rule loader and code-native
  presentation path already cover this scope.

## Deferred

- Sorcerer and Paladin higher-row skill additions.
- New skill capability types and more distinctive class mechanics.
- Broader skill-tree prerequisite restructuring.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestLoadRules|TestSkill|TestThirdClassSkillsRequirePrerequisites'`
- `.venv/bin/pytest tools/bot/test_skill_demo.py -q`
- `make client-unit`
- `make bot scenario=62_barbarian_earthbreaker.json`
- `make bot scenario=63_rogue_shadow_flurry.json`
- `make bot scenario=64_ranger_split_arrow.json`
- `make maintainability`
- `make ci`

# v59 As Built: Data-Driven Skill Catalog

Date: 2026-06-10
Spec: [`docs/specs/v59_spec-data-driven-skill-catalog.md`](../specs/v59_spec-data-driven-skill-catalog.md)
Plan: [`docs/plans/v59_2026-06-10-data-driven-skill-catalog.md`](../plans/v59_2026-06-10-data-driven-skill-catalog.md)

## What shipped

- Magic Bolt now lives in a schema-backed skill catalog with stable id `magic_bolt`, class metadata,
  tree placement, `projectile_attack` kind, targeting, max rank, `magic >= 15` requirement, bounded
  mana cost, rank-linear damage, projectile, and cooldown helpers.
- Skill presentation metadata is split into `shared/assets/skill_presentations.v0.json`, keeping
  client labels, icon swatches, summaries, and projectile visuals separate from server authority.
- The Go rules loader validates all catalog skills generically: supported kinds, helper types,
  requirements, prerequisite references, tree placement, projectile settings, cooldowns, and damage
  evaluators.
- Skill learning and casting share the same requirement helper. Attempts before `magic >= 15` reject
  with `skill_requirements_not_met` without spending points, mana, cooldowns, or combat state.
- Stat allocation now emits refreshed skill progression state, so the client sees Magic Bolt become
  learnable immediately after magic reaches 15.
- Godot now loads skill rules/presentations through `SkillRulesLoader`; the skill panel and skill
  bar render Magic Bolt name, requirement status, tooltip, slot text, and cooldown from shared data
  plus server progression/cooldown state.
- Protocol and client bot scenarios prove the clean-character loop: fail the Magic Bolt requirement,
  allocate enough magic, learn the skill, cast, reject cooldown recast, recover, and verify replay or
  fresh-session persistence.
- `docs/researchs/data-driven-content-libraries.md` records the v60-ready direction for content
  library manifests: file paths are loader indexes, while stable gameplay IDs remain the model.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run Skill`
- `cd server && go test ./internal/replay/... -run Skill`
- `cd server && go test ./internal/game/... ./internal/replay/... ./internal/http/...`
- `make client-unit`
- `make bot scenario=32_skill_points_and_magic_bolt.json`
- `SCENARIO=19_skill_points_and_magic_bolt.json ./scripts/bot_client_local.sh`
- `make ci`

## Deferred

- New active skills beyond Magic Bolt.
- Class selection, class-locked character creation, passive skills, respec/refund, and alternate
  skill branches.
- Free-form formula expressions and new skill capability types such as AoE, DOT, homing, summons,
  traps, auras, chained projectiles, and status effects.
- Production skill VFX, audio, and final combat balance.
- Full content library manifest rollout for items, skills, classes, and presentation assets.

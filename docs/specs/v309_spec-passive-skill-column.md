# v309 Spec — Passive Skill Column

Status: Complete
Date: 2026-06-20
Codename: passive-skill-column

## Purpose

Add a right-side passive skill column for every character class. Each class gets three one-rank passive skills in tree rows 1, 2, and 3, styled to the class identity, unlocked at character levels 1, 5, and 10, and chained so row 2 requires row 1 and row 3 requires row 2.

## Non-goals

- No respec/refund changes beyond the existing skill-rank reset behavior.
- No new active-skill casts, projectiles, status effects, animation events, or protocol message types.
- No external art, plugins, or generated bitmap logos; use existing code-native skill icon shapes and shared presentation metadata.
- No stat requirements for these passives.

## Class passive designs

| Class | Row 1, level 1 | Row 2, level 5 | Row 3, level 10 |
|-------|----------------|----------------|-----------------|
| Sorcerer | Arcane Focus — max mana | Mana Weaving — mana regeneration | Spell Dynamo — skill cooldown reduction |
| Barbarian | Iron Hide — max HP | Battle Tempo — attack speed | Crushing Force — weapon damage |
| Paladin | Vigilant Guard — armor | Faithful Bulwark — block chance | Consecrated Vitality — health regeneration |
| Rogue | Quick Hands — attack speed | Killer Instinct — critical chance | Evasive Footwork — evade chance |
| Ranger | Trail Sense — light radius | Precision Draw — hit chance | Deadeye — critical chance |

## Acceptance Criteria

- Each class has exactly three new passive nodes in a right-side tree column, one each on tiers 1, 2, and 3.
- The passive nodes have no base stat requirements and unlock at levels 1, 5, and 10 respectively.
- Row 2 passives require their class row 1 passive at rank 1, and row 3 passives require their row 2 passive at rank 1.
- The server applies learned passive bonuses to authoritative derived stats and stat breakdowns.
- Passive skills cannot be cast and do not consume mana or create cooldowns.
- Shared validation rejects missing passive presentations and verifies the new passive chain shape.
- Skill-panel tooltips show passive stat effects and code-native logos render from shared skill presentation data.

## Scope And Likely Files

- Shared rules/schema: `shared/rules/skills.v0.json`, `shared/rules/skills.v0.schema.json`, `shared/assets/skill_presentations.v0.json`, `shared/assets/skill_presentations.v0.schema.json`.
- Server authority: `server/internal/game/rules.go`, `server/internal/game/sim.go`, focused Go tests.
- Client presentation: `client/scripts/skills_panel.gd`, `client/tests/test_skills_panel.gd`.
- Validation/tooling: `tools/validate_skills.py`.
- Lifecycle docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, `docs/as-built/v309_passive-skill-column.md`.

## Test And Bot Proof

- `make validate-shared` proves schemas, presentations, and cross-checks.
- Focused Go tests prove passive unlock chains, class gates, non-castability, derived-stat effects, and breakdown sources.
- `make client-unit` proves the skill panel can display passive effects and icons.
- No bot scenario is required because this slice changes skill-tree progression and derived stats, with no new input flow beyond existing allocate/cast intents.

## Open Questions And Risks

- Existing skill-tree layout centers each row by visible sorted order; the new high-column nodes will appear to the right of existing row nodes without a layout rewrite.
- `server/internal/game/sim.go`, `server/internal/game/game_test.go`, and `client/scripts/skills_panel.gd` are grandfathered large files. Keep changes minimal and within their current baselines.

# v401 Spec — Class Skill Decuple Core

Status: Draft
Date: 2026-07-02
Codename: class-skill-decuple-core

## Purpose

Bring Barbarian, Paladin, Sorcerer, and Ranger to the **10-skill kit contract** (5 actives,
1 mobility, 4 passives). Remove redundant Ranger `split_arrow` and Sorcerer `arcane_orb`. Add
fourth passive column tier (level 15). Add missing Barbarian and Paladin actives/passives.
Rename shipped `ligthing` skill id to `lightning` with persistence alias migration.

## Non-goals

- Rogue `shadowstep` / `eviscerate` (v402).
- Skill-tree UI layout rewrite.
- Production VFX/audio; code-native presentations only.
- Final balance tuning across depths.

## Acceptance Criteria

- Each of barbarian, paladin, sorcerer, ranger has exactly 5 actives + 1 `mobility` + 4 passives.
- Barbarian gains `skullcrusher` (narrow cone nuke) and `unstoppable_heart` (tier-4 passive).
- Paladin gains `consecrated_smite` (cone AoE damage) and `oathbound_resolve` (tier-4 passive).
- Sorcerer gains `arcane_reservoir` (tier-4 passive); `arcane_orb` removed.
- Ranger gains `wildborn_endurance` (tier-4 passive); `split_arrow` removed.
- Tier-4 passives unlock at level 15, column 5, chained from tier-3 passive.
- `ligthing` renamed to `lightning`; saved `ligthing` ranks migrate to `lightning`.
- `make validate-shared` and focused class/skill tests pass.

## Scope And Likely Files

- `shared/rules/skills.v0.json`, `shared/assets/skill_presentations.v0.json`
- `shared/i18n/en.json`, `shared/i18n/es.json`
- `shared/content/codex_index.v0.json`, `shared/content/codex_overlays.v0.json`
- `server/internal/game/sim.go` (skill rank alias migration)
- `tools/validate_skills.py`
- Tests and bot scenarios referencing removed/renamed skills

## Test And Bot Proof

- `make validate-shared`
- Focused Go tests for new skills and passive tier 4
- Update extended scenarios using `ligthing` / `split_arrow`
- Extend barbarian/paladin foundation scenarios for new actives

## Open Questions And Risks

- `ligthing` rename touches many scenarios; alias migration must preserve persisted characters.
- Removing `split_arrow` retires scenario `64_ranger_split_arrow.json` (extended tier).

# v183 Plan: Ranger Black Wolf Companion

Status: Complete
Date: 2026-06-15
Spec: `docs/specs/v183_spec-ranger-wolf-companion.md`

## Adoption Checklist

- Decision: reject new plugin/asset dependency.
- Reason: the existing `monster_quadruped` scene, `dungeon_wolf` monster catalog pattern, and v182 companion entity path cover this slice. Authoritative summon/follow/attack remains in Go.
- Borrow/adopt: reuse existing monster visual catalog and client monster presentation code; add only data mappings required for black wolf tint/model.

## Tasks

- [x] Extend skill rules/schema/validation with a `summon_companion` kind and companion payload.
- [x] Add `black_wolf_companion` Ranger skill data and black-wolf companion monster/visual data.
- [x] Implement server summon handling: mana/cooldown, one active wolf per owner/skill, replacement, spawn change, and skill events.
- [x] Add focused Go tests for summon, replacement, owner, and black wolf view fields.
- [x] Add protocol bot scenario proving cast, spawn, follow, and companion damage.
- [x] Update docs/as-built and `PROGRESS.md`.
- [x] Run targeted checks: `make validate-shared`, focused Go tests, and `make bot scenario=74_ranger_wolf_companion.json`.
- [x] Run `make ci`.

## Bot Proof

Scenario: `tools/bot/scenarios/74_ranger_wolf_companion.json`

Expected flow:

1. Start a Ranger in a lab with seeded `black_wolf_companion` rank 1.
2. Cast the skill and observe `skill_cast` plus one living `companion` using the black wolf def.
3. Move the Ranger away and prove the wolf moves.
4. Wait for server AI to damage a nearby enemy with source entity type `companion`.

Visual verification command:

```bash
make bot-visual scenario=74_ranger_wolf_companion.json
```

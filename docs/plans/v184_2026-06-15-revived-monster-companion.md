# v184 Plan: Revived Monster Companion

Status: Complete
Date: 2026-06-15
Spec: `docs/specs/v184_spec-revived-monster-companion.md`

## Adoption Checklist

- Decision: reject new plugin/asset dependency.
- Reason: revived companions reuse the existing server companion AI and existing monster visual catalog identity.
- Borrow/adopt: use the current dead monster snapshot state (`hp: 0`) as the revive target and existing monster rendering for the spawned companion.

## Tasks

- [x] Add `revive_companion` skill kind, schema payload, and rules validation.
- [x] Add Sorcerer `revive` shared skill data, presentation, and localized text.
- [x] Implement targeted revive handling with dead-target, living-target, and boss rejection.
- [x] Spawn revived monsters as owned companions with rank-scaled HP/damage and original monster identity.
- [x] Add focused Go tests for revive, scaling, and boss/living rejection.
- [x] Extend bot target resolution for dead monster entities.
- [x] Add protocol bot scenario proving kill, revive, companion spawn, and companion damage.
- [x] Update docs/as-built and `PROGRESS.md`.
- [x] Run `make ci`.

## Bot Proof

Scenario: `tools/bot/scenarios/75_sorcerer_revive_companion.json`

Expected flow:

1. Start a Sorcerer in a revive lab with seeded `revive` rank 1.
2. Kill a non-boss `dungeon_wolf`.
3. Cast Revive on the dead wolf entity.
4. Observe one living `companion` using `dungeon_wolf` monster identity.
5. Wait for companion-sourced damage against `combat_lab_soft_target`.

Visual verification command:

```bash
make bot-visual scenario=75_sorcerer_revive_companion.json
```

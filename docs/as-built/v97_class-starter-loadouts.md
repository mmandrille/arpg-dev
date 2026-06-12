# v97 As-Built — Class starter loadouts

Date: 2026-06-12
Status: Complete

## What Shipped

- Added shared starter templates for paladin sword/shield, sorcerer staff, and barbarian axe, with placeholder visual and presentation mappings.
- Seeded explicit character creation with durable common starter equipment plus one `red_potion` and one `blue_potion`.
- Added `max_mana` and `skill_damage_percent` item stats, including validation, pricing schema coverage, derived max-mana aggregation, and skill-damage scaling from equipped items.
- Added a protocol bot scenario proving a newly created sorcerer starts with the staff equipped, empty offhand, and both potions in inventory.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/http -run 'TestCharacterClassSeedsSessionStartProgression|TestCreatedCharactersReceiveClassStarterLoadouts'`
- `cd server && go test ./internal/game -run TestStarterStaffAddsMaxManaAndSkillDamage`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make bot scenario=46_class_starter_loadout`
- `make maintainability`
- `make ci`

## Scope Limits

- Starter loadouts are created only on explicit character creation. Existing characters and login-created compatibility defaults are not backfilled.
- Starter equipment reuses placeholder item visuals; dedicated axe, staff, and shield production art remains deferred.
- `skill_damage_percent` affects skill damage resolution but is not exposed as a dedicated top-level derived-stat display yet.

## Maintainability Note

This slice touched grandfathered creation, simulation, and bot-test files in narrow paths. Starter-specific HTTP coverage and skill-stat helpers were extracted to focused files, while the remaining `sim.go` max-mana/stat-breakdown growth is covered by a documented baseline update. The baseline also records the already-shipped v96 `client/scripts/main.gd` growth and current `auth_session_test.go` size so the ratchet is accurate again.

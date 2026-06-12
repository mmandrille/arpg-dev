# v97 Spec — Class starter loadouts

Status: Complete
Date: 2026-06-12
Codename: `class-starter-loadouts`

## Purpose

Newly created heroes start with a class-appropriate common equipment package and one health plus one mana potion. Paladins start with a one-handed sword and shield, sorcerers start with a two-handed staff that grants a little max mana and can roll magic caster bonuses, and barbarians start with a slower but harder-hitting two-handed axe.

The loadout is durable server-owned inventory state, so fresh sessions, resume snapshots, bot scenarios, and the Godot client all observe the same starter items without client-side shortcuts.

## Non-goals

- Backfilling already-created characters, including login-created compatibility defaults.
- New production art, models, icons, or a Godot equipment plugin.
- New class selection UI; existing character creation/class picker stays in place.
- A full affix naming system or elemental skill-scaling redesign beyond the starter staff stat needed here.

## Acceptance Criteria

- Creating a barbarian persists a common two-handed axe equipped in `main_hand`, leaves `off_hand` empty, and adds one `red_potion` plus one `blue_potion`.
- Creating a sorcerer persists a common two-handed staff equipped in `main_hand`, leaves `off_hand` empty, grants base `max_mana`, and supports magic rolled `mana_regen_per_10_seconds` plus `skill_damage_percent`.
- Creating a paladin persists a common one-handed sword in `main_hand`, a shield in `off_hand`, and adds one `red_potion` plus one `blue_potion`.
- Starter items are loaded through the existing session-start snapshot and inventory/equipment protocol fields; no protocol version bump is required.
- Shared rules validation accepts the new item templates, stat key, and presentation mappings.
- Focused tests prove starter loadouts for all three classes and prove equipped `skill_damage_percent` increases skill damage deterministically.
- Bot proof covers at least one created class seeing the starter equipped items and potions from the same protocol path used by players.

## Scope and Likely Files

- `shared/rules/item_templates.v0.json` and schema: starter templates, new `axe`/`staff` item types, and `skill_damage_percent`.
- `shared/assets/item_visuals.v0.json` and `shared/assets/item_presentations.v0.json`: placeholder visual/presentation mappings for starter template ids.
- `server/internal/http/character.go` or a small adjacent helper: seed starter inventory when a character is explicitly created.
- `server/internal/game/sim.go` / tests: apply equipped `skill_damage_percent` to skill damage.
- `server/internal/http/auth_session_test.go`, `server/internal/game/*_test.go`, and bot scenario/test files: coverage.
- `docs/plans/`, `docs/as-built/`, and `PROGRESS.md`: lifecycle close-out.

## Test and Bot Proof

- `make validate-shared`
- Focused Go tests for character creation/session loadout seeding.
- Focused Go test for `skill_damage_percent` on a rolled staff.
- Bot scenario proving a sorcerer start has the starter staff equipped and both potions in inventory.
- Final `make maintainability` and `make ci`.

## Open Questions and Risks

- Existing characters are intentionally not backfilled; if needed later, this should be a migration/backfill slice.
- Starter equipment reuses existing placeholder visuals; richer axe/staff/shield art remains a visual polish follow-up.

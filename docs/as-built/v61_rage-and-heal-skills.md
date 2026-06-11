# v61 As Built: Rage and Heal Skills

Date: 2026-06-10
Spec: [`docs/specs/v61_spec-rage-and-heal-skills.md`](../specs/v61_spec-rage-and-heal-skills.md)
Plan: [`docs/plans/v61_2026-06-10-rage-and-heal-skills.md`](../plans/v61_2026-06-10-rage-and-heal-skills.md)

## What shipped

- Added `rage` and `heal` to the first skill-tree row alongside `magic_bolt`.
- Replaced projectile-only skill schema assumptions with a closed declarative `effects` model:
  `stat_percent_buff` for Rage and `area_percent_heal` for Heal. JSON cannot dispatch arbitrary
  plugin/method names.
- Rage costs 10 mana, requires STR 10 and VIT 10 plus 5 per extra rank, applies a 450-tick
  STR/VIT percent buff, emits start/end effect events, updates max HP while active, and sets
  player `visual_scale` so the character and equipped gear scale together.
- Heal costs 10 mana, requires MAGIC 10 plus 5 per extra rank, resolves its area from a target
  entity or cast direction, heals allied living players including the caster, clamps to missing HP,
  and emits `player_healed` events using the existing green floating text path.
- Protocol v8 event schemas now allow skill casts without `projectile_def_id` and skill-sourced
  `player_healed` events without potion item IDs.
- The Godot skills panel renders/selects all first-row skills, spends points by selected skill id,
  and exposes per-skill debug state for client bot assertions.
- The skill bar now tracks the selected/right-click skill instead of assuming the first catalog row.
  Self and area skills send target-shaped payloads when appropriate; Magic Bolt projectile behavior
  remains unchanged.
- Added a protocol bot scenario, `rage_and_heal_skills`, that learns and casts Rage, then proves
  Heal in a fresh `heal_lab` session using persisted skill ranks.
- Added a compact `skill_progression_lab` so Magic Bolt progression proofs no longer depend on long
  generated-dungeon traversal.
- Stabilized affected combat/replay/client smoke proofs by forcing hit-chance paths only in tests
  that require deterministic hits, retrying scenario actions through legal misses, and removing
  incidental player-damage assumptions from equip/resume smoke coverage.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game/...`
- `cd server && go test ./...`
- `make client-unit`
- `make bot SCENARIO=rage_and_heal_skills`
- `make bot`
- `make bot-client SCENARIO=model_reaction_polish HEADLESS=1`
- `make ci`

## Deferred

- Explicit ground-position targeting for Heal.
- Buff icons, timers, and richer skill HUD state.
- Production Rage/Heal VFX/audio beyond scale and existing heal popups.
- Party/team semantics beyond allied player entities.
- Additional effect types, passive skills, free-form formulas, or user-authored executable plugins.

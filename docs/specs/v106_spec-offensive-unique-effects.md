# v106 Spec: Offensive Unique Effects

Status: Complete
Date: 2026-06-12
Codename: `offensive-unique-effects`

## Purpose

Make the offensive unique effects in `shared/rules/unique_effects.v0.json` mechanically live:
`stormbound_echo`, `executioners_mark`, and `hunger_of_the_deep`.

This slice turns the catalog from passive data into authoritative combat behavior for direct
damage procs while keeping all tuning in shared rules. It builds on v105's live unique-effect
pipeline and should leave reusable helpers for later defensive and support unique effects.

## Player-Facing Behavior

- `stormbound_echo`: basic attacks have the configured chance to chain lightning damage to one
  nearby enemy. The secondary hit uses the configured percent of the triggering hit and
  `lightning` damage type.
- `executioners_mark`: damaging a monster at or below the configured low-health threshold marks it.
  If that monster dies before the mark expires, nearby monsters take a configured physical damage
  pulse based on the marking hit.
- `hunger_of_the_deep`: repeated hits against the same monster build stacking bonus damage. Stacks
  expire after the configured idle window and reset when the hero damages a different monster.

## Non-Goals

- No defensive, resource, support, or movement unique effects; those are planned for v107/v108.
- No new protocol schema version. Existing `monster_damaged`, `monster_killed`,
  `skill_effect_started`, and `skill_effect_ended` events are sufficient.
- No new Godot visual effects. Client presentation can show existing damage/floating text; bespoke
  lightning, mark, or stack visuals are deferred.
- No balance pass on unique drop rates or item rarity weights.

## Data And Authority

- The Go sim owns all effect triggering, RNG rolls, damage, stacks, marks, expiration, and kill
  credit.
- Values come from `shared/rules/unique_effects.v0.json`.
- `stormbound_echo` uses the deterministic sim RNG. It must not use wall-clock time or unseeded
  randomness.
- Effects must only trigger from equipped item `effect_ids`.
- The client remains presentation-only.

## Acceptance Criteria

- Go tests prove `stormbound_echo` triggers from a basic attack at the configured chance and emits a
  secondary `monster_damaged` event with `skill_id: "stormbound_echo"` and `damage_type:
  "lightning"`.
- Go tests prove `stormbound_echo` does not trigger from skill damage.
- Go tests prove `executioners_mark` starts on low-health damaged monsters, expires
  deterministically, and pulses nearby enemies if the marked monster dies in time.
- Go tests prove `hunger_of_the_deep` increases repeated same-target damage, resets on target
  change, and expires after the configured idle window.
- Protocol bot coverage proves a unique-equipped hero can trigger at least one offensive unique
  effect through the normal WebSocket path.
- `make validate-shared`, focused Go tests, bot proof, maintainability, and final `make ci` pass.

## Scope And Likely Files

- `server/internal/game/unique_effects.go` — offensive unique effect state and triggers.
- `server/internal/game/sim.go` — player snapshot fields and hook calls around direct damage.
- `server/internal/game/types.go` — event fields only if existing fields are insufficient.
- `server/internal/game/unique_effects_test.go` — focused deterministic tests.
- `shared/rules/worlds.v0.json` and/or bot scenario JSON — lab proof for offensive unique procs.
- `tools/bot/scenarios/...` — scenario coverage for normal protocol path.
- `PROGRESS.md`, `docs/as-built/v106_offensive-unique-effects.md`.

## Test And Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestUniqueBurn|TestOffensiveUnique|TestUniqueEffect'`
- `ARPG_BOT_SCENARIO=offensive_unique_effects VERBOSE=1 make bot`
- `make maintainability`
- `make ci`

## Open Questions And Risks

- No blocking questions. Default: use current catalog tuning as written.
- Main risk: `server/internal/game/sim.go` and `game_test.go` are already large. Prefer new helper
  structs/functions in `unique_effects.go` and focused tests in `unique_effects_test.go`.
- Main determinism risk: proc chance and target search. Use seeded `RNG` and sorted entity ids.

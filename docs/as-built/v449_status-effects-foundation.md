# v449 — Status effects foundation

## What it proved

- Added a shared `status_effects.v0.json` catalog for reusable poison, burn, and slow debuffs.
- New poison DOT, unique burn DOT, weapon burn DOT, weapon poison DOT, and ice slow applications now flow through one authoritative server status map.
- Status ticks stay deterministic, resistance-aware, and emit the existing `skill_effect_started` / `monster_damaged` / `skill_effect_ended` events.
- Authoritative slow percent now feeds monster movement and attack-speed reads.
- Cleanse removes debuff-tagged player status entries when present.
- Client poison/burn/slow presentation now keys off authoritative effect IDs instead of single-skill special cases.
- Focused proof passed with protocol bot scenarios `undead_skeleton_poison_immunity`, `unique_burn_effect_live`, `status_effect_slow_live`, and client bot scenario `unique_burn_effect_live`.

## Key decisions

- Keep protocol v8 unchanged and reuse existing combat/status event shapes.
- Leave legacy bleed/root/stun paths in place; this slice migrates new poison/burn/slow producers first.
- Add optional bot runtime `max_distance` support for future movement-budget assertions, but keep the live slow scenario event-based because spawn-to-end movement is not a sound proxy for post-application slow strength.

## Deferred

- Full migration of bleed/root/stun onto the same framework.
- Player-targeted poison/burn/slow applications and a live bot cleanse scenario once those states can occur through regular gameplay.
- Broader status bar/icon treatment beyond current marker/tint coverage.

## Verification

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestWeaponElementalProc|TestOffensiveUniqueReplicatingBlight|TestRogueMarkIncreasesPoisonTickDamage|TestSurvivalUniqueFrostglassWardSlowsAndBuffsAfterLargeHit|TestSurvivalUniqueAshenReprisalPrimesAndConsumesOnNextHit|TestSurvivalEvasiveStanceCleanseRemovesStatusDebuffs' -count=1
make bot scenario=undead_skeleton_poison_immunity
make bot scenario=unique_burn_effect_live
make bot scenario=status_effect_slow_live
make bot-client SCENARIO=unique_burn_effect_live HEADLESS=1
```

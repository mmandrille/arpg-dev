# v449 Plan — Status Effects Foundation

Status: Ready for implementation
Goal: Move new slow/poison/burn applications onto one data-driven status framework without a protocol bump.
Architecture: shared status catalog -> server-owned status map/tick loop -> client effect-id presentation -> focused Go/bot/client proofs.

## Baseline and decisions

- Reuse existing bleed/root/stun paths where they already work; migrate only new poison/burn/slow applications in this slice.
- Keep v8 protocol in place and preserve existing `skill_effect_started` / `monster_damaged` event shapes.
- Adopt in-repo marker/tint assets only; no external plugins.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/rules/status_effects.v0.json` | Status catalog |
| Create | `shared/rules/status_effects.v0.schema.json` | Schema |
| Modify | `server/internal/game/rules.go` | Load + validate status rules |
| Create | `server/internal/game/status_effects.go` | Shared status state/tick helpers |
| Modify | `server/internal/game/rogue_skills.go` | Poison DOT migration |
| Modify | `server/internal/game/weapon_elemental_procs.go` | Burn/poison/slow proc migration |
| Modify | `server/internal/game/unique_effects.go`, `unique_debuff_replication.go`, `unique_survival_effects.go` | Unique burn/slow hooks |
| Modify | `server/internal/game/movement.go`, `sim.go`, `sim_players.go`, `tick_results.go`, `survival_skills.go` | Persistence, movement reads, cleanse, tick order |
| Modify | `client/scripts/player_status_effect_markers.gd`, `client/scripts/main.gd` | Effect-id presentation |
| Modify | `tools/bot/runtime_assertions.py` | Optional movement-cap assertions for future status labs |
| Create | `tools/bot/scenarios/115_status_effect_slow_live.json` | Slow live proof |
| Modify | `shared/rules/worlds.v0.json` | Slow lab world |

## Tasks

## Task 1 — Shared catalog

- [x] Add status rules/schema and server validation

## Task 2 — Server framework

- [x] Add authoritative status state, DOT ticking, slow queries, persistence, and cleanup
- [x] Migrate new poison/burn/slow producers onto the framework
- [x] Cover cleanse with a focused server test

## Task 3 — Client presentation

- [x] Move poison/burn/slow presentation to authoritative effect IDs

## Task 4 — Bot/client proof

- [x] Reuse poison and burn extended scenarios
- [x] Add slow live protocol scenario
- [x] Re-run unique burn client presentation proof

## Final verification

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestWeaponElementalProc|TestOffensiveUniqueReplicatingBlight|TestRogueMarkIncreasesPoisonTickDamage|TestSurvivalUniqueFrostglassWardSlowsAndBuffsAfterLargeHit|TestSurvivalUniqueAshenReprisalPrimesAndConsumesOnNextHit|TestSurvivalEvasiveStanceCleanseRemovesStatusDebuffs' -count=1
make bot scenario=undead_skeleton_poison_immunity
make bot scenario=unique_burn_effect_live
make bot scenario=status_effect_slow_live
make bot-client SCENARIO=unique_burn_effect_live HEADLESS=1
```

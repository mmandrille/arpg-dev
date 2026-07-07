# v449 Spec: Status Effects Foundation

Status: Approved for implementation
Date: 2026-07-07
Codename: `status-effects-foundation`
Baseline: v448 `quest-steward-hunt-loop`

## Purpose

Ship one authoritative debuff model for reusable **slow**, **poison DOT**, and **burn DOT** effects so branch uniques, weapon procs, and future skill hooks stop inventing ad hoc status state.

## Non-goals

- Full crowd-control catalog (stun, fear, root, silence)
- PvP balance, production VFX/audio, protocol bump beyond v8
- Full migration of every legacy debuff in one slice

## Acceptance criteria

- Shared `status_effects.v0.json` owns supported debuff IDs, kind, damage type, affected stats, and debuff flag.
- Server loads/validates the catalog and applies new poison/burn/slow effects through one status-effect state map.
- New DOT ticks are deterministic, duration-driven, resistance-aware, and emit the same authoritative damage events clients/bots already consume.
- Slow effects affect authoritative monster movement and attack-speed reads.
- Cleanse removes debuff-tagged player status entries when present.
- Client poison/burn/slow presentation keys off authoritative effect IDs instead of single-skill special cases.
- Focused proof covers poison, burn, slow, and cleanse using Go tests plus extended bot/client scenarios.

## Scope and files

- Shared: `shared/rules/status_effects.v0.json`, schema, world/scenario additions
- Server: status effect state + loaders, rogue poison, weapon elemental procs, unique burn/slow hooks, cleanse integration
- Client: `player_status_effect_markers.gd`, `main.gd`
- Bot: slow live scenario, existing poison/burn extended proofs

## Verification

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestWeaponElementalProc|TestOffensiveUniqueReplicatingBlight|TestRogueMarkIncreasesPoisonTickDamage|TestSurvivalUniqueFrostglassWardSlowsAndBuffsAfterLargeHit|TestSurvivalUniqueAshenReprisalPrimesAndConsumesOnNextHit|TestSurvivalEvasiveStanceCleanseRemovesStatusDebuffs' -count=1
make bot scenario=undead_skeleton_poison_immunity
make bot scenario=unique_burn_effect_live
make bot scenario=status_effect_slow_live
make bot-client SCENARIO=unique_burn_effect_live HEADLESS=1
```

## Asset decision

- Adopt: existing in-repo status marker/tint presentation and bot labs
- Reject: external plugins/assets

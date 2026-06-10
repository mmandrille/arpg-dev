# v56 Plan - Monster Attack Cadence

Status: Ready for implementation
Goal: Slightly increase generated dungeon monster attack cadence without changing damage,
movement, boss timing, or protocol shape.
Architecture: This is a shared-rules tuning slice with test ownership. The Go server remains
authoritative and already reads monster cooldowns from `shared/rules/monsters.v0.json`; the client
has no gameplay authority and needs no code path for this slice.
Tech stack: Shared JSON rules/goldens, Go sim tests, GDScript golden parity, Python protocol bot,
SDD docs.

## Baseline and shortcut decision

Baseline is v55 `consolidation-and-quality-gates` on `main`. Reuse existing monster-rule loading,
proactive monster attack tests, ranged archer projectile tests, and bot scenarios.

Godot plugin adoption: not applicable. This slice has no client UI/art/camera/presentation work.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/monsters.v0.json` | Tune generated dungeon melee/ranged attack cooldowns |
| Modify | `shared/golden/dungeon_monster_attack.json` | Record tuned melee cooldown ownership |
| Modify | `shared/golden/dungeon_monster_attack.v0.schema.json` | Validate cooldown field |
| Modify | `shared/golden/item_rolls.json` | Restore required description metadata found by shared validation |
| Modify | `server/internal/game/game_test.go` | Assert golden cooldown and existing cooldown behavior |
| Modify | `client/tests/test_golden.gd` | Cross-check golden cooldown against shared rules |
| Add | `docs/specs/v56_spec-monster-attack-cadence.md` | Slice spec |
| Add | `docs/plans/v56_2026-06-10-monster-attack-cadence.md` | This plan |
| Modify | `PROGRESS.md` | Lifecycle close-out |
| Add | `docs/as-built/v56_monster-attack-cadence.md` | As-built proof |

## Task 1 - Shared rule tuning

Files:
- Modify: `shared/rules/monsters.v0.json`

- [x] Step 1.1: Change `dungeon_mob.attack_cooldown_ticks` from `40` to `32`.
- [x] Step 1.2: Change `dungeon_archer.attack_cooldown_ticks` from `90` to `75`.
- [x] Step 1.3: Validate shared data.
- [x] Step 1.4: Restore missing `shared/golden/item_rolls.json` description metadata exposed by
  the shared validation gate.

```bash
make validate-shared
```

## Task 2 - Golden ownership

Files:
- Modify: `shared/golden/dungeon_monster_attack.json`
- Modify: `shared/golden/dungeon_monster_attack.v0.schema.json`
- Modify: `server/internal/game/game_test.go`
- Modify: `client/tests/test_golden.gd`

- [x] Step 2.1: Add `attack_cooldown_ticks` to the dungeon monster attack golden fixture and schema.
- [x] Step 2.2: Assert the fixture cooldown equals the loaded `dungeon_mob` rule in Go.
- [x] Step 2.3: Assert the fixture cooldown equals the loaded `dungeon_mob` rule in GDScript.
- [x] Step 2.4: Keep the existing "no attack before cooldown, attack after cooldown" Go proof green.

```bash
cd server && go test ./internal/game/... -run 'TestDungeonMonster(ProactiveAttackGolden|AttackCooldownAndDeterminism)' -count=1
make client-unit
```

## Task 3 - Ranged archer proof

Files:
- Modify if needed: `server/internal/game/game_test.go`
- Existing: `tools/bot/scenarios/38_ranged_monster_ai.json`

- [x] Step 3.1: Run the focused archer projectile test against the tuned cooldown.
- [x] Step 3.2: Run the existing protocol bot archer scenario.

```bash
cd server && go test ./internal/game/... -run TestRangedMonsterProjectileDamagesPlayer -count=1
make bot scenario=38_ranged_monster_ai.json
```

## Task 4 - Regression bot gates

Files:
- Existing: `tools/bot/scenarios/24_boss_floor_gate.json`
- Existing: `tools/bot/scenarios/32_skill_points_and_magic_bolt.json`

- [x] Step 4.1: Run boss-floor gate because it traverses multiple generated floors and boss combat.
- [x] Step 4.2: Run skill-point/magic-bolt because it kills multiple generated dungeon mobs.

```bash
make bot scenario=24_boss_floor_gate.json
make bot scenario=32_skill_points_and_magic_bolt.json
```

## Task 5 - Lifecycle docs and CI

Files:
- Modify: `docs/plans/v56_2026-06-10-monster-attack-cadence.md`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v56_monster-attack-cadence.md`

- [x] Step 5.1: Mark plan checkboxes complete after focused gates pass.
- [x] Step 5.2: Add v56 lifecycle/as-built docs and record deferred balance scope.
- [x] Step 5.3: Run full CI.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'TestDungeonMonster(ProactiveAttackGolden|AttackCooldownAndDeterminism)|TestRangedMonsterProjectileDamagesPlayer' -count=1`
- [x] `make client-unit`
- [x] `make bot scenario=38_ranged_monster_ai.json`
- [x] `make bot scenario=24_boss_floor_gate.json`
- [x] `make bot scenario=32_skill_points_and_magic_bolt.json`
- [x] `make ci`

## Deferred scope

- Final combat balance pass across damage, HP, movement, rarity, and depth scaling.
- Boss pattern cadence and boss UI readability, queued separately as selected autoloop slices.

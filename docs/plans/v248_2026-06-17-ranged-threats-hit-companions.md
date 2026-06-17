# v248 Plan - Ranged Threats Hit Companions

Status: Complete
Goal: Let ranged monsters damage engaged companions via projectiles.
Architecture: Reuse existing engaged companion target selection, route ranged monster projectiles
to that target, and resolve targeted companion projectile hits through the existing companion damage
path.
Tech stack: Go sim, shared world fixture, protocol bot, docs.

## Baseline and Asset Decision

Builds on v182 companion AI, v208 stance targeting, v220 mercenary loss, and v247 companion combat
stats. No client visual changes are required.

Asset/plugin decision:
- Adopt existing projectile presentation and combat events.
- Borrow existing melee engaged-companion targeting.
- Reject external assets/plugins and new projectile visuals.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/sim.go` | Route ranged projectiles to selected target and include companion hits |
| Add | `server/internal/game/ranged_companion_projectile_test.go` | Focused Go proof |
| Modify | `shared/rules/worlds.v0.json` | Add compact ranged companion bot lab |
| Add | `tools/bot/scenarios/91_ranged_threats_hit_companions.json` | Protocol bot proof |
| Add | `docs/as-built/v248_ranged-threats-hit-companions.md` | As-built proof |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines, except grandfathered baselines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`

Decision:
- [x] Keep new test coverage in a new focused file.
- [x] Keep `sim.go` growth small and avoid unrelated cleanup.

Verification:
```bash
make maintainability
```

## Task 1 - Server targeting and projectile impact

Files:
- Modify: `server/internal/game/sim.go`
- Add: `server/internal/game/ranged_companion_projectile_test.go`

- [x] Let ranged monster attacks call `fireMonsterProjectile` with the selected monster attack target.
- [x] Include targeted companions in monster-owned projectile hit detection.
- [x] Resolve companion projectile hits through `damageCompanionByMonster`.
- [x] Prove projectile target ID and companion damage/death events.

```bash
cd server && go test ./internal/game -run 'RangedMonster.*Companion|RangedMonsterProjectile' -count=1
```

## Task 2 - Bot fixture

Files:
- Modify: `shared/rules/worlds.v0.json`
- Add: `tools/bot/scenarios/91_ranged_threats_hit_companions.json`

- [x] Add a compact lab with an archer, companion, player, and walls.
- [x] Add a protocol bot scenario that waits for archer-sourced `companion_damaged` or
  `attack_missed` against a companion.

```bash
make validate-shared
make bot scenario=91_ranged_threats_hit_companions.json
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v248_ranged-threats-hit-companions.md`

- [x] Record focused verification and deferred scope.

## Final Verification

- [x] `cd server && go test ./internal/game -run 'RangedMonster.*Companion|RangedMonsterProjectile' -count=1`
- [x] `make validate-shared`
- [x] `make bot scenario=91_ranged_threats_hit_companions.json`
- [x] `make maintainability`

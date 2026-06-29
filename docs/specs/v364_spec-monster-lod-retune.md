# v364 Spec: Monster Movement LOD Retune

Status: Draft  
Date: 2026-06-28  
Codename: `monster-lod-retune`  
Baseline: v363 `basic-attack-cooldown-tuning`

## Purpose

Retune data-driven monster movement LOD thresholds so crowded combat stays within navigation
budgets while near-player monsters remain responsive. Builds on v270 crowd movement LOD and v269
navigation budgets.

## Non-goals

- No LOD logic rewrites or new overload policies.
- No client presentation changes.
- No protocol or golden formula changes beyond navigation snapshot fixtures.

## Acceptance criteria

- `shared/rules/navigation.v0.json` LOD fields retuned with documented rationale in as-built.
- Crowded lightning probe still produces both LOD-deferred and high-precision monsters.
- Focused Go tests prove LOD remains inactive in small fights and elites/bosses stay high precision.
- New semantic test proves retuned near-distance keeps a meaningful high-precision share in crowded
  rooms without exceeding configured path budgets.

## Scope and files

| Area | Files |
|------|-------|
| Rules | `shared/rules/navigation.v0.json` |
| Goldens | `shared/golden/monster_chase.json`, `shared/golden/auto_path.json` |
| Server tests | `server/internal/game/monster_movement_lod_test.go` |
| Docs | spec, plan, as-built, lifecycle, `PROGRESS.md` |

## Test proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'MovementLOD|CrowdedLightning' -count=1
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
```

## Asset decision

N/A — server/rules-only slice.

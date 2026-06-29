# v365 Spec: Path Budget Retune

Status: Draft  
Date: 2026-06-28  
Codename: `path-budget-retune`  
Baseline: v364 `monster-lod-retune`

## Purpose

Retune authoritative navigation path budgets so crowded combat spends fewer path nodes per tick while
monsters still move and path counters stay within configured caps.

## Non-goals

- No client changes.
- No new overload degradation policies.
- No pathfinding algorithm changes.

## Acceptance criteria

- `shared/rules/navigation.v0.json` monster path request/node budgets retuned.
- Existing crowded budget tests pass.
- Semantic test asserts average per-tick path nodes stay below a rule-derived ceiling during crowded
  probe warmup.

## Test proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'CrowdedLightning|PathBudget' -count=1
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
```

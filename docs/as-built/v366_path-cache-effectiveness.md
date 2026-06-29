# v366 As-Built — Path Cache Effectiveness

## What shipped

- `monster_path_cache_ticks`: 8 → **12**; `monster_repath_throttle_ticks`: 6 → **8**.
- New `monster_path_cache_goal_tolerance`: **1.0** world units for chase-goal reuse.
- `monsterPathGoalMatchesCached` helper in `monster_navigation_budget.go`.
- Test: cache hits ≥ 90% of path requests over 32 crowded ticks.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'CrowdedLightning' -count=1
```

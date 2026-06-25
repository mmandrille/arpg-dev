# v336 — Wall-floor lab navigation proof

**Status:** Complete  
**Codename:** wall-floor-lab-nav

## What it proved

- `server/internal/game/wall_floor_lab_nav_test.go` proves `generated_wall_lab` stairs-down offset move goals remain reachable after navigation fixes.
- Client bot scenario `tools/bot/scenarios/client/78_wall_floor_shader_polish.json` exercises wall layout across a stairs-down transition (v308 presentation path).

## Verification

```bash
cd server && go test ./internal/game/... -run TestGeneratedWallLabStairsOffsetMoveGoal
HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=wall_floor_shader_polish ./scripts/bot_visual.sh
```

## Deferred

- Production wall-floor art pass beyond code-native shaders.

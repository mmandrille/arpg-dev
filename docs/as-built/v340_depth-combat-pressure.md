# v340 — Depth combat pressure

Raised data-driven `monster_depth_scaling` knobs and updated `monster_rarity` golden expectations. Added `monster_depth_pressure_test.go`.

Verification: `cd server && go test ./internal/game/... -run TestGeneratedMonsterStatsDepthPressure`, `make validate-shared`

# v63 Plan - Runtime Sim Error Construction

Status: Implemented (`make ci` green)
Goal: Replace the default exported sim constructor panic path with an error-returning API.
Architecture: Keep all gameplay construction behavior identical. `NewSim` becomes the default-world
error-returning constructor, matching `NewSimWithWorld` and `NewSimWithWorldProgression`.
Tests can use an explicit `MustNewSim` helper, while runtime code must handle returned errors.
Tech stack: Go sim constructors and tests, lifecycle docs.

## Baseline and shortcut decision

Baseline is v62 `monster-depth-stat-scaling` on `main`. This slice is backend-only and does not
not applicable.

Bot scenarios are not required because the slice changes internal constructor failure handling only;
`make ci` still runs the full bot suite as the final gate.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `docs/specs/v63_spec-runtime-sim-error-construction.md` | Slice spec |
| Add | `docs/plans/v63_2026-06-11-runtime-sim-error-construction.md` | This plan |
| Modify | `server/internal/game/sim.go` | `NewSim` error return plus explicit `MustNewSim` helper |
| Modify | `server/internal/game/game_test.go` | Update test setup calls |
| Modify | `server/internal/game/shop_test.go` | Update test setup calls |
| Modify | `server/internal/realtime/session_loop_test.go` | Update test setup call |
| Modify | `tools/bot/scenarios/12_dungeon_levels.json` | Remove stale generated gold assertions exposed by CI |
| Modify | `tools/bot/test_protocol.py` | Align scenario unit assertion with travel-only dungeon-level contract |
| Modify | `client/tests/test_golden.gd` | Align dungeon stairs golden check with current shared fixture |
| Modify | `PROGRESS.md` | Lifecycle status and deferred backlog |
| Add | `docs/as-built/v63_runtime-sim-error-construction.md` | As-built summary |

## Task 1 - Constructor API

Files:
- Modify: `server/internal/game/sim.go`

- [x] Step 1.1: Change `NewSim(sessionID, seed, rules)` to return `(*Sim, error)` from
  `NewSimWithWorld(..., DefaultWorldID)`.
```bash
cd server && go test ./internal/game/... -run TestNewSim -count=1
```

- [x] Step 1.2: Add `MustNewSim(sessionID, seed, rules)` for tests and local helpers that
  intentionally want panic-on-invalid-fixture behavior.
```bash
cd server && go test ./internal/game/... -run TestNewSim -count=1
```

- [x] Step 1.3: Add a focused test proving invalid default-world construction returns an error
  from `NewSim` instead of panicking.
```bash
cd server && go test ./internal/game/... -run TestNewSim -count=1
```

## Task 2 - Test Call Sites

Files:
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/game/shop_test.go`
- Modify: `server/internal/realtime/session_loop_test.go`

- [x] Step 2.1: Update existing test setup call sites from `NewSim` to `MustNewSim` where the
  loaded fixtures are expected to be valid.
```bash
cd server && go test ./internal/game/... ./internal/realtime/... -count=1
```

- [x] Step 2.2: Update any call sites that should assert constructor errors to use the
  error-returning `NewSim` directly.
```bash
rg -n "NewSim\\(" server -g'*.go'
cd server && go test ./... -count=1
```

## Task 3 - Lifecycle Docs And CI

Files:
- Modify: `docs/plans/v63_2026-06-11-runtime-sim-error-construction.md`
- Modify: `tools/bot/scenarios/12_dungeon_levels.json`
- Modify: `tools/bot/test_protocol.py`
- Modify: `client/tests/test_golden.gd`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v63_runtime-sim-error-construction.md`

- [x] Step 3.1: Remove stale generated gold expectations from `dungeon_levels`, align the Python
  scenario unit test, and align the GDScript dungeon stairs golden check with the current
  `shared/golden/dungeon_stairs.json` fixture. This scenario now owns stair traversal only, while
  v49 gold scenarios own currency pickup behavior.
```bash
make bot scenario=dungeon_levels
```

- [x] Step 3.2: Mark completed plan steps, add v63 lifecycle/as-built docs, and record the
  closed v60 review finding.
```bash
rg -n "v63|runtime-sim-error-construction|Latest completed slice|NewSim" PROGRESS.md docs/as-built docs/plans/v63_2026-06-11-runtime-sim-error-construction.md
```

- [x] Step 3.3: Run final verification.
```bash
make test-go
make ci
```

## Final verification

- [x] `make test-go`
- [x] `make ci`

## Deferred scope

- Broader constructor cleanup outside sim startup remains future refactor work if another panic
  wrapper appears in runtime code.

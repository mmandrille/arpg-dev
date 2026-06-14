# v142 As-built: Sim Load And Players Extraction

Spec: [`docs/specs/v142_spec-sim-load-and-players-extraction.md`](../specs/v142_spec-sim-load-and-players-extraction.md)
Plan: [`docs/plans/v142_2026-06-13-sim-load-and-players-extraction.md`](../plans/v142_2026-06-13-sim-load-and-players-extraction.md)

## What shipped

- Moved persisted item, hotbar, skill binding, stash, shop stock, and teleporter load methods into
  `server/internal/game/sim_load.go`.
- Moved payload parsing/clone/string helpers with the load-adjacent package utilities.
- Moved co-op guest creation, player metadata/connectivity/query helpers, removal/respawn, town
  spawn selection/blocking helpers, active-player context switching, player save/default helpers,
  and deterministic player ID sorting into `server/internal/game/sim_players.go`.
- Lowered the `server/internal/game/sim.go` maintainability baseline from 7801 to 7045 lines.
- Updated CODEMAP so session/realtime work points at the focused sim load/player files.

## Proof

- `cd server && go test ./internal/game -run 'TestLoadInventoryAppliesEquippedResourceStats' -count=1`
- `cd server && go test ./internal/game ./internal/realtime -run 'Test.*(Guest|Player|Respawn|Coop|SessionLoop)' -count=1`
- `make lint-determinism`
- `cd server && go test ./internal/game ./internal/realtime`
- `.venv/bin/python tools/validate_codemap.py`
- `make maintainability`
- `make ci`

## Deferred

- Broader `game_test.go` domain split remains deferred.
- Further tick-loop/combat extraction from `sim.go` remains deferred to future gameplay slices.

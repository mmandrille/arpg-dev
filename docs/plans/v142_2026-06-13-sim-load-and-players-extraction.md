# v142 Plan: Sim Load And Players Extraction

Status: Ready for implementation
Goal: Split persistence-load and player lifecycle helpers out of `sim.go` without behavior changes.
Architecture: This is a Go package-internal refactor in `server/internal/game`. The deterministic
tick loop and input handling stay in `sim.go`; persistence loading moves to `sim_load.go`, while
player lifecycle, spawn, and active-player context helpers move to `sim_players.go`.
Tech stack: Go sim, determinism lint, replay-sensitive Go tests, maintainability ratchet.

## Baseline and shortcut decision

Builds on v141 `market-store-extraction` and the v140 review recommendation to reduce large
coordinators. No Godot/client shortcut decision is required; this slice does not touch client UI,
camera, inventory presentation, or art.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `server/internal/game/sim_load.go` | Persisted load methods and payload clone helpers. |
| Create | `server/internal/game/sim_players.go` | Player lifecycle, spawn, context switching, and sorted player IDs. |
| Modify | `server/internal/game/sim.go` | Remove moved code and stale imports. |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `sim.go` grandfathered baseline after extraction. |
| Modify | `docs/CODEMAP.md` | Point relevant domains at the split sim files. |
| Create | `docs/as-built/v142_sim-load-and-players-extraction.md` | Close-out proof and deferred scope. |
| Modify | `PROGRESS.md` | Mark v142 complete and update current status. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 - Extract load helpers

Files:
- Create: `server/internal/game/sim_load.go`
- Modify: `server/internal/game/sim.go`

- [x] Step 1.1: Move persisted load structs and `Load*` methods to `sim_load.go`.
- [x] Step 1.2: Move payload clone/string helper functions with the load block, preserving package visibility.
- [x] Step 1.3: Remove stale imports from `sim.go`.
```bash
gofmt -w server/internal/game/sim.go server/internal/game/sim_load.go
cd server && go test ./internal/game -run 'TestLoadInventoryAppliesEquippedResourceStats' -count=1
```

## Task 2 - Extract player lifecycle helpers

Files:
- Create: `server/internal/game/sim_players.go`
- Modify: `server/internal/game/sim.go`

- [x] Step 2.1: Move `AddGuestPlayer`, metadata/connectivity/query helpers, remove/respawn, town spawn, player context, save, and sort helpers to `sim_players.go`.
- [x] Step 2.2: Preserve deterministic iteration ordering and existing `//nolint:determinism` comments where needed.
- [x] Step 2.3: Run focused co-op/respawn coverage.
```bash
gofmt -w server/internal/game/sim.go server/internal/game/sim_players.go
cd server && go test ./internal/game ./internal/realtime -run 'Test.*(Guest|Player|Respawn|Coop|SessionLoop)' -count=1
```

## Task 3 - Ratchet, CODEMAP, lifecycle, and CI

Files:
- Modify: `.maintainability/file-size-baseline.tsv`
- Modify: `docs/CODEMAP.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v142_sim-load-and-players-extraction.md`
- Modify: `docs/specs/v142_spec-sim-load-and-players-extraction.md`
- Modify: `docs/plans/v142_2026-06-13-sim-load-and-players-extraction.md`

- [x] Step 3.1: Lower the `sim.go` baseline to the post-extraction line count.
- [x] Step 3.2: Update CODEMAP and lifecycle docs.
- [x] Step 3.3: Run final verification.
```bash
make lint-determinism
cd server && go test ./internal/game ./internal/realtime
make maintainability
make ci
```

## Final verification

- [x] `make lint-determinism`
- [x] `cd server && go test ./internal/game ./internal/realtime`
- [x] `make maintainability`
- [x] `make ci`

## Deferred scope

- Broader `game_test.go` domain split remains deferred.
- Further tick-loop/combat extraction from `sim.go` remains deferred to future gameplay slices.

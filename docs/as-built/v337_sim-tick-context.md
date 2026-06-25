# v337 — Sim tick context extraction

**Status:** Complete  
**Codename:** sim-tick-context

## What it proved

- `server/internal/game/sim_tick_context.go` introduces typed `simTickCtx` bundling per-tick result accumulation for `TickResultsProfiled`.
- `tick_results.go` delegates result-key bookkeeping to the context without changing tick semantics.

## Verification

```bash
cd server && go test ./internal/game/... -run TestStopMovement
cd server && go test ./internal/game/... -count=1
```

## Deferred

- Multi-slice `sim.go` coordinator paydown using the context for phase helpers.

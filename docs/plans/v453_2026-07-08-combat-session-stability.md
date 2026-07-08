# v453 Plan — Combat Session Stability

Spec: [`docs/specs/v453_spec-combat-session-stability.md`](../specs/v453_spec-combat-session-stability.md)

## Tasks

- [x] Store: `AppendEvents` batch insert
- [x] `persist_batch.go`: batch event persist; extend defer ops (resource bag/wallet)
- [x] `session_tick.go`: defer on total/persist over-budget; coalesced fanout
- [x] `fanout_coalesce.go` + blocking `enqueue`
- [x] Unit tests: defer trigger, persist batch, enqueue backpressure
- [x] Extended bot scenario `ranger_volley_session_stability`
- [x] PROGRESS + as-built

## Final verification

```bash
cd server && go test ./internal/realtime/... -count=1
ARPG_PERF_DEBUG=1 make bot scenario=ranger_volley_session_stability
```

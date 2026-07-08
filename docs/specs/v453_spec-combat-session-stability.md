# v453 Spec — Combat Session Stability

Status: Approved  
Date: 2026-07-08  
Codename: `combat-session-stability`

## Purpose

Stop skill-cast WebSocket disconnects in solo crowded combat by fixing persist/fanout backpressure. No skill semantics or visual changes.

## Non-goals

- `skill_damage_burst` event aggregation (v454)
- Skill resolution model / projectile entity removal
- Multiplayer soak (v456)
- Async persist worker

## Acceptance criteria

1. `deferNonCritical` arms when total tick or persist phase exceeds budget, not sim-only.
2. Session events persist in one batch per tick result; resource-bag/wallet writes participate in defer/batch policy.
3. Same-tick `state_delta` envelopes coalesce per client; full `sendCh` blocks/coalesces instead of hard-closing WS.
4. Extended scenario `ranger_volley_session_stability`: Volley ×5 in crowded lab, WS stays connected.
5. `cd server && go test ./internal/realtime/... -count=1` green.

## Scope

| Area | Files |
|------|-------|
| Tick defer | `server/internal/realtime/session_tick.go` |
| Persist batch | `server/internal/realtime/persist_batch.go`, `persist_defer.go` |
| Fanout / enqueue | `server/internal/realtime/fanout_coalesce.go`, `session_loop.go` |
| Store | `server/internal/store/repos.go`, `interfaces.go` |
| Bot | `tools/bot/scenarios/*ranger_volley_session_stability*` |
| Tests | `session_tick_test.go`, `session_loop_test.go` |

## Test proof

```bash
ARPG_PERF_DEBUG=1 make bot scenario=ranger_volley_session_stability
cd server && go test ./internal/realtime/... -count=1
```

## Open questions

None — batch clarification answered in autoloop brief.

# v453 As-Built — combat-session-stability

Date: 2026-07-08  
Spec: [`docs/specs/v453_spec-combat-session-stability.md`](../specs/v453_spec-combat-session-stability.md)  
Plan: [`docs/plans/v453_2026-07-08-combat-session-stability.md`](../plans/v453_2026-07-08-combat-session-stability.md)

## Shipped

- `deferNonCritical` arms on sim, total elapsed, or persist-phase budget overrun (not sim-only).
- Session events batch-persist via `AppendEvents` in one transaction per tick result.
- Resource wallet/bag change ops participate in defer policy (`OpResourceWalletUpdate`, `OpResourceBagItemAdd`).
- Same-tick `state_delta` coalesced per client via `fanoutTickResults`.
- Full `sendCh` blocks instead of hard-closing WebSocket connections.
- Extended scenario `ranger_volley_session_stability`: Volley ×5 in crowded lab without transport drop.

## Verification

```bash
cd server && go test ./internal/realtime/... -count=1
ARPG_PERF_DEBUG=1 make bot scenario=ranger_volley_session_stability
```

## Deferred

- `skill_damage_burst` aggregation (v454).
- Skill resolution hybrid migration (v454).
- Per-tick skill budget scheduler (v455).

# v457 As-Built — live combat transport stability

Date: 2026-07-08

Spec: [`docs/specs/v457_spec-live-combat-transport-stability.md`](../specs/v457_spec-live-combat-transport-stability.md)

Plan: [`docs/plans/v457_2026-07-08-live-combat-transport-stability.md`](../plans/v457_2026-07-08-live-combat-transport-stability.md)

## Finding

The reported disconnects did not coincide with sustained backend tick pressure: surrounding ticks
were approximately 3–11 ms against a 100 ms budget. The earlier v453–v456 proofs used Python peers
or offline Godot replay, so they bypassed the failing live Godot WebSocket/frame-processing path.

## Shipped

- Data-driven Godot WebSocket capacity: 4 MiB inbound, 256 KiB outbound, 8,192 queued packets,
  and a five-second heartbeat.
- Client close code/reason diagnostics and backend read/write error diagnostics.
- A lifetime reconnect counter exposed to client-bot assertions.
- `ranger_volley_live_session_stability`: five Volley casts in the crowded lab through a live
  Godot connection, with a zero-reconnect assertion.
- Class-aware client-bot character selection and class-preserving debug progression setup.
- Performance testing guidance distinguishing protocol, replay, observer, and authoritative-client
  test topologies.
- Final CI recovery for preceding unpushed work: aligned quest-leaf proofs with resource-bag
  ownership, registered missing scenario-audit rows, and sorted anchor-cluster map keys to remove a
  flaky dungeon-layout golden caused by Go map iteration.

## Proof

```bash
ARPG_PERF_DEBUG=1 SCENARIO=ranger_volley_live_session_stability HEADLESS=1 ./scripts/bot_client_local.sh
make validate-shared
cd server && go test ./internal/realtime/... ./internal/http/... -count=1
godot --headless --path client --script res://tests/test_net_client.gd
godot --headless --path client --script res://tests/test_connection_recovery_runtime.gd
```

The live scenario passed in 22.91 seconds with zero reconnects. The owner also confirmed repeated
interactive Ranger Volley play worked without reconnects.

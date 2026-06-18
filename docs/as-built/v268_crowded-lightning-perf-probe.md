# v268 As-Built - Crowded Lightning Perf Probe

Date: 2026-06-18
Spec: [`docs/specs/v268_spec-crowded-lightning-perf-probe.md`](../specs/v268_spec-crowded-lightning-perf-probe.md)
Plan: [`docs/plans/v268_2026-06-18-crowded-lightning-perf-probe.md`](../plans/v268_2026-06-18-crowded-lightning-perf-probe.md)

## Shipped Behavior

- Backend perf debug logs now split sampled tick cost into `ai_ms`, `pathfind_ms`, `combat_ms`,
  `broadcast_ms`, and `persist_ms`, plus tick budget/overrun fields.
- The deterministic sim records per-tick `path_requests`, `path_cache_hits`,
  `path_nodes_visited`, and `monsters_moved` without importing wall-clock time into `game/`.
- The crowded lightning probe world stages 36 server-owned chase monsters in a vault-like room and
  a sorcerer with `ligthing` unlocked through debug progression.
- `crowded_lightning_perf_probe` repeatedly casts `ligthing` and is available as both protocol bot
  and Godot visual replay.
- `HEADLESS=1 make bot-visual ...` now maps to Godot `--headless` for replay scenarios.

## Perf Sample

The protocol probe emitted backend rows that localize the current cost to navigation/pathfinding.
One sampled crowded tick reported:

```text
tick=138 total_ms=37.731 sim_ms=36.134 ai_ms=36.066 pathfind_ms=34.937 path_requests=202 path_nodes_visited=5825 monsters_moved=4
```

This is the baseline for v269 navigation budgeting.

## Boundaries

- Monster navigation remains server-authoritative.
- No path reuse, repath throttling, movement LOD, or overload degradation shipped in v268.
- Client work is limited to replay/debug flag plumbing; this is not a client visual-only fix.
- `.maintainability/file-size-baseline.tsv` lowers `sim.go` to the post-extraction count and records
  current pre-existing client baseline drift for `client/scripts/main.gd` and
  `client/tests/test_coop_client.gd` so the ratchet gate is green again.

## Usage

```bash
make play-debug
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
```

`make play-debug` still captures interactive perf output in `/tmp/arpg-perf.log`.

## Verification

```bash
make validate-shared
cd server && go test ./internal/game ./internal/realtime
.venv/bin/pytest tools/bot/test_protocol.py -q
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
make maintainability
```

All focused commands passed on 2026-06-18. Final full `make ci` remains the enclosing autoloop batch gate.

# v268 Plan - Crowded Lightning Perf Probe

Status: Complete
Goal: Make crowded lightning freezes localizable with backend phase timings, path counters, and a repeatable bot/visual stress scenario.
Architecture: Keep the authoritative sim deterministic and timing-free by exposing counters from
`game/` while realtime owns wall-clock measurements. Reuse v267 opt-in `ARPG_PERF_DEBUG` sampling
and `make play-debug`; add one crowded lab world and one scenario that can run as protocol bot and
visual replay. This slice observes cost only; navigation budgeting and degradation are deferred.
Tech stack: Go sim/realtime, shared JSON rules, Python protocol bot, Godot visual replay wrapper, SDD docs.

## Baseline and shortcut decision

Builds on v267 `perf-debug-mode`. No external assets or plugins are needed; the visual proof
borrows the existing protocol-bot recording plus Godot replay path and existing skill/projectile
presentation.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/pathfind.go` | Add deterministic path search stats without timing |
| Modify | `server/internal/game/perf_debug.go` | Expose per-tick path/movement counters |
| Modify | `server/internal/game/sim.go` | Wrap tick phases and route sim path requests through counters |
| Add | `server/internal/game/tick_results.go` | Keep tick processing extracted from the shrinking sim coordinator |
| Modify | `server/internal/game/elite_minion_ai.go` | Route monster pathing through sim path counters |
| Modify | `server/internal/game/companion_ai.go` | Route companion pathing through sim path counters |
| Modify | `server/internal/game/approach.go` | Route player auto-nav pathing through sim path counters |
| Modify | `server/internal/game/handlers.go` | Route move-to pathing through sim path counters |
| Modify | `server/internal/realtime/perf_debug.go` | Add backend perf log fields and phase profiler |
| Modify | `server/internal/realtime/session_loop.go` | Measure sim, persist, broadcast, and phase timings |
| Add | `server/internal/realtime/session_tick.go` | Keep session tick processing out of the shrinking loop coordinator |
| Modify | `shared/rules/monsters.v0.json` | Add no-drop crowded-probe chaser monster |
| Modify | `shared/rules/worlds.v0.json` | Add crowded lightning stress world |
| Modify | `shared/i18n/en.json` | Add probe monster display name |
| Modify | `shared/i18n/es.json` | Add probe monster display name |
| Add | `tools/bot/scenarios/93_crowded_lightning_perf_probe.json` | Reproduce crowded lightning combat |
| Modify | `tools/bot/test_protocol.py` | Assert scenario discovery and stress shape |
| Modify | `make/agents.mk` | Pass `HEADLESS=1` through visual bot replay |
| Modify | `scripts/bot_local.sh` | Pass `ARPG_PERF_DEBUG` to bot server |
| Modify | `scripts/bot_visual.sh` | Pass `ARPG_PERF_DEBUG` to visual bot server and replay client |
| Modify | `.maintainability/file-size-baseline.tsv` | Lock in `sim.go` shrink and refresh pre-existing client drift |
| Add | `docs/as-built/v268_crowded-lightning-perf-probe.md` | Record shipped proof |
| Modify | `docs/progress/slice-lifecycle.md` | Add v268 lifecycle row |
| Modify | `docs/progress/scenario-catalog.md` | Add scenario catalog row |
| Modify | `PROGRESS.md` | Advance status after focused verification |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd` was not touched; baseline refreshed to current pre-existing count after the ratchet blocked the gate.
- [x] `server/internal/game/game_test.go` is not touched
- [x] `tools/bot/run.py` is not touched
- [x] `tools/validate_shared.py` is not touched
- [x] Other over-limit files from `.maintainability/file-size-baseline.tsv`: `server/internal/game/sim.go` (shrunk), `server/internal/realtime/session_loop.go` (shrunk), `tools/bot/test_protocol.py` (within allowance), `client/tests/test_coop_client.gd` (not touched; baseline refreshed to current pre-existing count).
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice, or
- [ ] Defer extraction with rationale: the `sim.go` edit is a narrow call-site instrumentation
  wrapper inside the existing tick loop; extracting gameplay phases before measuring them would
  blur the perf-probe slice and risk behavior drift.

Extraction note: tick processing moved to `server/internal/game/tick_results.go` and realtime tick
processing moved to `server/internal/realtime/session_tick.go`, lowering `sim.go` to 6488 lines and
`session_loop.go` to 899 lines.

Verification:

```bash
make maintainability
```

## Task 1 - Backend counters and phase timings

Files:
- Modify: `server/internal/game/pathfind.go`
- Modify: `server/internal/game/perf_debug.go`
- Modify: `server/internal/game/sim.go`
- Add: `server/internal/game/tick_results.go`
- Modify: `server/internal/game/elite_minion_ai.go`
- Modify: `server/internal/game/companion_ai.go`
- Modify: `server/internal/game/approach.go`
- Modify: `server/internal/game/handlers.go`
- Modify: `server/internal/realtime/perf_debug.go`
- Modify: `server/internal/realtime/session_loop.go`
- Add: `server/internal/realtime/session_tick.go`

- [x] Add path search stats and per-tick counters for requests, cache hits, nodes visited, and monsters moved.
- [x] Add game tick phase profiler hooks without wall-clock imports in `game/`.
- [x] Measure `ai_ms`, `pathfind_ms`, `combat_ms`, `broadcast_ms`, `persist_ms`, tick budget, and overrun fields in realtime.
- [x] Keep default non-debug behavior unchanged.

```bash
cd server && go test ./internal/game ./internal/realtime
```

## Task 2 - Crowded lightning reproduction

Files:
- Modify: `shared/rules/worlds.v0.json`
- Modify: `shared/rules/monsters.v0.json`
- Modify: `shared/i18n/en.json`
- Modify: `shared/i18n/es.json`
- Add: `tools/bot/scenarios/93_crowded_lightning_perf_probe.json`
- Modify: `tools/bot/test_protocol.py`
- Modify: `make/agents.mk`
- Modify: `scripts/bot_local.sh`
- Modify: `scripts/bot_visual.sh`

- [x] Add a vault-like crowded world with 30-40 live chase monsters.
- [x] Add the `crowded_lightning_perf_probe` scenario with sorcerer debug progression and repeated `ligthing` casts.
- [x] Pass `ARPG_PERF_DEBUG` through bot local/visual server paths so the command emits backend perf logs.
- [x] Add scenario discovery/shape test coverage.

```bash
make validate-shared
.venv/bin/pytest tools/bot/test_protocol.py -q
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
```

## Task 3 - Lifecycle docs

Files:
- Add: `docs/as-built/v268_crowded-lightning-perf-probe.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/progress/scenario-catalog.md`
- Modify: `PROGRESS.md`
- Modify: `docs/specs/v268_spec-crowded-lightning-perf-probe.md`
- Modify: `docs/plans/v268_2026-06-18-crowded-lightning-perf-probe.md`

- [x] Mark spec and plan complete after focused verification.
- [x] Record exact commands and the manual visual command.
- [x] Advance `PROGRESS.md` to v268 with final batch CI still pending for autoloop.

```bash
make maintainability
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game ./internal/realtime`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe`
- [x] `ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe`
- [x] `make maintainability`

Final full `make ci` is deferred to the enclosing `$autoloop` batch gate.

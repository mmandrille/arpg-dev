# v268 Spec - Crowded Lightning Perf Probe

Status: Complete
Date: 2026-06-18
Codename: crowded-lightning-perf-probe

## Purpose

Make crowded lightning-combat freezes diagnosable before changing navigation behavior. Extend the
v267 perf debug mode so local `make play-debug` and an automated bot/visual scenario can show
whether 300-600ms live ticks are dominated by AI movement, pathfinding, combat, broadcast fanout,
or persistence.

## Non-goals

- No navigation optimization, path reuse, throttling, LOD, or overload degradation in this slice.
- No client-authoritative monster movement or client-hosted navigation.
- No protocol/schema change; perf proof stays in logs and bot scenario data.
- No production metrics backend or dashboard.
- No final combat or monster density tuning.

## Acceptance Criteria

- `make play-debug` continues to capture opt-in perf output to `/tmp/arpg-perf.log`.
- Backend `backend_perf` logs include the existing v267 fields plus `ai_ms`, `pathfind_ms`,
  `combat_ms`, `broadcast_ms`, `persist_ms`, `path_requests`, `path_cache_hits`,
  `path_nodes_visited`, `monsters_moved`, `tick_budget_ms`, `tick_over_budget`, and
  `tick_overrun_ms`.
- The sim exposes deterministic pathfinding counters without using wall-clock time in `game/`;
  realtime code owns wall-clock phase measurement.
- A new crowded lightning world preset has 30-40 live chase monsters, vault-like walls, and a
  sorcerer player with `ligthing` unlocked through debug progression.
- A protocol bot scenario repeatedly casts `ligthing` in that world and proves the scenario
  reaches live crowded combat while perf debug logging can localize tick cost.
- A visual replay command exists for the same scenario so the freeze case is repeatable:
  `ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe`.
- Focused tests cover backend perf field presence, path counter population, shared-rule validity,
  and bot scenario discovery.

## Scope and Likely Files

- Server:
  - `server/internal/game/pathfind.go`
  - `server/internal/game/perf_debug.go`
  - `server/internal/game/sim.go`
  - `server/internal/game/tick_results.go`
  - `server/internal/game/elite_minion_ai.go`
  - `server/internal/game/companion_ai.go`
  - `server/internal/game/approach.go`
  - `server/internal/game/handlers.go`
  - `server/internal/realtime/perf_debug.go`
  - `server/internal/realtime/session_loop.go`
  - `server/internal/realtime/session_tick.go`
- Shared rules:
  - `shared/rules/monsters.v0.json`
  - `shared/rules/worlds.v0.json`
  - `shared/i18n/en.json`
  - `shared/i18n/es.json`
- Bot/tooling:
  - `tools/bot/scenarios/93_crowded_lightning_perf_probe.json`
  - `tools/bot/test_protocol.py`
  - `make/agents.mk`
  - `scripts/bot_local.sh`
  - `scripts/bot_visual.sh`
- Docs:
  - `docs/plans/v268_2026-06-18-crowded-lightning-perf-probe.md`
  - `docs/as-built/v268_crowded-lightning-perf-probe.md`
  - `docs/progress/slice-lifecycle.md`
  - `docs/progress/scenario-catalog.md`
  - `PROGRESS.md`

## Test and Bot Proof

```bash
make validate-shared
cd server && go test ./internal/game ./internal/realtime
.venv/bin/pytest tools/bot/test_protocol.py -q
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
make maintainability
```

Manual visual proof command:

```bash
make play-debug
```

Use the crowded lightning world manually through the new bot/visual scenario first; `make
play-debug` remains the human interactive path with `/tmp/arpg-perf.log`.

## Open Questions and Risks

- No blocking product questions. The requested typo spelling `ligthing` is preserved because the
  current shared skill id uses that spelling.
- The crowded scenario may need a higher per-scenario elapsed budget than normal because it
  intentionally creates stress. Keep the budget explicit in the scenario rather than weakening the
  global bot budget.
- `game/` must not import or call wall-clock APIs; all timing duration measurement is owned by
  realtime.

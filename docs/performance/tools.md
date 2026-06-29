# Performance tools reference

Canonical guide for **how this repo instruments, samples, and stress-tests performance**.
Use this when investigating stutter, tick overruns, or render pressure — and when **adding**
new metrics, probes, or debug hooks.

**Not** a post-mortem for a specific incident. Record those under `docs/performance/investigations/`
(optional) or slice as-builts under `docs/as-built/`.

**Code index:** [`docs/CODEMAP.md`](../CODEMAP.md) → row **Performance debug**.

---

## Quick start

| Goal | Command |
|------|---------|
| Interactive play with perf logs (tee to file) | `make play-debug` → `/tmp/arpg-perf.log` |
| **Analyze a real play session** | `make perf-analyze LOG=/tmp/arpg-perf.log` |
| Same, explicit env | `ARPG_PERF_DEBUG=1 make play` |
| Protocol bot + backend perf | `ARPG_PERF_DEBUG=1 make bot scenario=<id>` |
| Godot replay + perf | `ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=<id>` |
| In-game overlay (no log parsing) | Settings → **Performance status** (client panel) |
| **Full benchmark suite + report** | `make benchmark` |

### How `make benchmark` works — live concurrent session

The benchmark runs the **protocol bot and Godot client simultaneously on the same live session**:

1. Server starts with `ARPG_PERF_DEBUG=1`.
2. For each benchmark scenario, the bot creates a listed co-op session and writes the session ID to a temp file.
3. Three seconds later, Godot launches and joins the same session as a second player (`ARPG_JOIN_SESSION_ID`).
4. The bot drives the scenario (combat, spells, movement). Godot renders what it sees in real-time.
5. When the bot finishes, Godot is closed. The server log and Godot's stdout are combined into the report.

This captures real `[client-perf]` FPS under actual server load — not a replay. The CLIENT section of
the report shows what a second player sees while the first player (bot) is actively fighting.

**To measure your own play session:** use `make play-debug`, reproduce the slow scenario for 2–3 minutes,
then `make perf-analyze LOG=/tmp/arpg-perf.log`.

### What to look for in real play logs

The `perf-analyze` report CLIENT section calls out the metrics that explain low FPS:

| Metric | What it means when high |
|--------|------------------------|
| `p5 fps` | Worst-tail FPS — the "20 fps" you're feeling |
| `fog ms` | Fog-of-war shader update cost; grows with dungeon area revealed |
| `d_upsert_player` | Local player state upsert; spikes when inventory/quest/reconciliation triggers |
| `d_chg` | Full entity change loop; grows with entity count in delta |
| `draw_calls` | Scene complexity; high in dense dungeons with many wall segments |
| `d_recon` | Player reconciliation backpressure; high when input lags behind server |

Correlate spikes: when `p5 fps` drops, check which phase is highest on the same sample line.

**Master switch:** `ARPG_PERF_DEBUG=1` (or `true` / `yes` / `on`). Wired through `scripts/play.sh`,
`scripts/bot_local.sh`, `scripts/bot_visual.sh`, and `scripts/benchmark.sh` to **both** Go server
and Godot client. Default play is unchanged when unset.

**Correlate client vs server:** Backend healthy + client `[client-perf]` bad → client presentation.
Backend `tick_over_budget` / high `sim_ms` + client fine → sim, pathfind, persist, or fanout.

---

## Client tools (Godot)

### `[client-perf]` log lines

- **Sampler:** `client/scripts/perf_debug_sampler.gd` — ~1 Hz while `ARPG_PERF_DEBUG` is on.
- **Hook:** `main.gd` `_process()` calls `_perf_debug_sampler.sample(...)`.

Each line includes: `fps`, `avg_frame_ms`, `process_ms`, `physics_ms`, `tick`, WebSocket state,
`recon_delta`, entity counts, Godot node/object/draw-call/primitive counts, then **phase suffix**
when phases were recorded that second.

Example shape:

```text
[client-perf] fps=58 avg_frame_ms=17.2 ... tick=1204 ... delta=45.12 d_ui=12.30 d_chg=28.50 ...
```

Phase values are **milliseconds accumulated in that 1s window** (sum across frames), not single-frame latency.

### `PerfPhaseTimer` — phase buckets

- **File:** `client/scripts/perf_phase_timer.gd` (`class_name PerfPhaseTimer`)
- **API:** `ensure_enabled()`, `measure_usec(phase, start_usec)`, `format_snapshot(rank_by_value)`, `reset_frame()`
- **Tests:** `client/tests/test_perf_phase_timer.gd`

Enabled only when `ARPG_PERF_DEBUG` is set. Zero overhead when off.

#### Registered phases (maintain this table when adding hooks)

| Phase | Location | Meaning |
|-------|----------|---------|
| `net_poll` | `main.gd` `_process()` | WebSocket poll / message dequeue |
| `entities` | `main.gd` `_process()` | Entity tick smoothing |
| `fog` | `fog_of_war_overlay.gd` | Fog overlay update |
| `delta` | `main.gd` `_apply_delta()` | Total authoritative delta apply |
| `d_prep` | `_apply_delta()` | Level-change prep; mobility index from events |
| `d_chg` | `_apply_delta()` | `changes[]` loop |
| `d_upsert` | `_upsert_entity()` | All entity upserts (subset of `d_chg`) |
| `d_upsert_m` | `_upsert_entity()` | Monster upserts only |
| `d_upsert_player` | `_upsert_entity()` | Local/authoritative player upserts only |
| `d_ui` | `_apply_delta()` | Throttled inventory / quest / minimap sync |
| `d_evt` | `_apply_delta()` | `events[]` presentation |
| `d_dfog` | `_apply_delta()` | Fog wall resync when needed |
| `d_bot` | `_apply_delta()` | Bot event tagging |
| `d_boss` | `_apply_delta()` | Boss health bar sync |
| `d_recon` | `_apply_delta()` | `_reconcile_player()` |

**Adding a client phase**

1. Pick a short `snake_case` name; prefix delta sub-phases with `d_`.
2. Wrap with `var t := Time.get_ticks_usec()` … `PerfPhaseTimerScript.measure_usec("name", t)`.
3. Document the row in the table above.
4. Add or extend `test_perf_phase_timer.gd` if format/aggregation behavior changes.
5. Append a row to [Changelog](#changelog) below.

### Performance status overlay (in-game)

- **Wire payload:** `state_delta.performance` (optional) — see [Backend tools](#backend-tools-go).
- **Formatter:** `client/scripts/performance_status_formatter.gd`
- **Storage:** `main.gd` → `last_performance_status`; shown when user enables **Performance status** in settings.
- **Slice:** v272 — [`docs/as-built/v272_performance-status-overlay.md`](../as-built/v272_performance-status-overlay.md)

Shows FPS, client ping estimate, backend `total_ms` / `sim_ms` / phase splits, path counters,
room shape, loop counts, tick budget / overrun / degradation flag.

### Client presentation load-shedding (not metrics — affects perf)

Data-driven caps/throttles agents should know about when interpreting FPS:

| Mechanism | Config | File(s) |
|-----------|--------|---------|
| Entity presentation LOD | `shared/rules/main_config.v0.json` → `presentation_lod` | `entity_presentation_lod.gd` |
| Projectile visible cap | `client_perf.projectile_visible_cap` | `projectile_presentation_cap.gd` |
| Loot label crowd cull | `loot_labels.*` | `loot_label_filter.gd` |
| Fog combat shader throttle | `shared/assets/fog_presentation.v0.json` | `fog_of_war_overlay.gd` |
| Delta frame coalesce | (code) | `delta_frame_coalesce.gd` |
| Delta UI sync gate | `client_perf.delta_ui_sync_interval_ticks`, `delta_minimap_sync_interval_ticks` | `delta_ui_sync_gate.gd` |
| Reconciliation backpressure | `client_perf.reconciliation_backpressure_threshold` | `reconciliation_backpressure.gd` |
| Windup marker cap | `client_perf.windup_marker_max_concurrent` | `monster_melee_windup_marker.gd` |

Loader: `client/scripts/main_config_loader.gd`. Schema: `shared/rules/main_config.v0.schema.json`.

---

## Backend tools (Go)

### `backend_perf` structured logs

- **Emitter:** `server/internal/realtime/perf_debug.go` → `logBackendPerf`
- **Interval:** ~1s when `ARPG_PERF_DEBUG` is set
- **Tests:** `server/internal/realtime/perf_debug_test.go`

Fields include: `tick`, `total_ms`, `sim_ms`, `ai_ms`, `pathfind_ms`, `combat_ms`,
`broadcast_ms`, `persist_ms`, `path_requests`, `path_cache_hits`, `path_nodes_visited`,
`monsters_moved`, `tick_budget_ms`, `tick_over_budget`, `tick_overrun_ms`, `inputs`, `results`,
`changes`, `events`, `acks`, `rejects`, `clients`, `game_level`, entity breakdown, `walls`.

Look for log lines with `"msg":"backend_perf"` (prefixed `[backend]` in `make play`).

### `state_delta.performance` payload

Same shape as logs, fanout to clients on a throttled **performance-only** delta (empty
`changes` / `events`) for the in-game overlay. Built by `buildPerformanceStatus()` in
`perf_debug.go`. Schema: `shared/protocol/state_delta.v*.schema.json` → `performance` object.

### Deterministic work counters (replay-safe)

- **File:** `server/internal/game/perf_debug.go`
- **Struct:** `PerfCounters` — `PathRequests`, `PathCacheHits`, `PathNodesVisited`, `MonstersMoved`
- **Reset:** `resetTickPerf()` each tick (also resets navigation budgets + collision cache)

Use counters for **regression tests** and bot assertions; use wall-clock fields for **local profiling**.

### Tick phase profiler (wall-clock, realtime only)

- **Interface:** `game.TickProfiler` → `MeasureTickPhase(name, fn)`
- **Implementation:** `backendTickProfiler` in `realtime/perf_debug.go`
- **Phase names:** `game.TickPhaseAI`, `TickPhaseCombat`, `TickPhasePathfind`

### Tick guardrails & load shedding

| Module | Role |
|--------|------|
| `server/internal/realtime/tick_guardrails.go` | 10 Hz budget evaluation, overrun ms |
| `server/internal/game/combat_tick_budget.go` | Combat-phase movement throttle |
| `server/internal/game/monster_overload_guardrails.go` | Overload degradation policy |
| `server/internal/game/persist_defer.go` | Defer non-critical DB writes when sim over budget |
| `server/internal/game/tick_collision_cache.go` | Per-tick collision cache |

Tuning: `shared/rules/navigation.v0.json` (`monster_overload_*`, path budgets). Some budgets remain
code-owned (e.g. combat phase ms) — check file before assuming data-driven.

### Benchmark report generator

- **File:** `tools/bot/benchmark_report.py`
- **Called by:** `scripts/benchmark.sh` (do not invoke directly unless debugging)
- **Inputs:** `--server-log` (Go structured JSON) + `--bot-log` (bot stderr with scenario markers)
- **Output:** per-scenario summary using `backend_perf` lines; slices by wall-clock timestamp from bot markers

**Adding a backend metric**

1. Prefer **deterministic counters** in `perf_debug.go` for replay/bot use; wall-clock only in `realtime/`.
2. Extend `buildPerformanceStatus` + `logBackendPerf` + protocol `performanceStatusPayload` together.
3. Bump protocol schema if wire shape changes.
4. Update `performance_status_formatter.gd` if user-visible.
5. Add Go test in `realtime/perf_debug_test.go` or focused `*_test.go`.
6. Append [Changelog](#changelog).

---

## Bot scenarios & lab worlds (stress / regression)

Reproducible perf paths without manual dungeon walks. Two tiers:

- **`ci_tier: extended`** — included in `make ci-full`, excluded from `make ci`. Manual run with `make bot scenario=<id>`.
- **`ci_tier: benchmark`** — excluded from all CI. Only launched with `make benchmark`, which runs all benchmark scenarios, opens the visual Godot client with the Performance status overlay, and emits a report.

### Extended probes

| Scenario ID | File | Focus |
|-------------|------|-------|
| `crowded_lightning_perf_probe` | `tools/bot/scenarios/93_crowded_lightning_perf_probe.json` | Crowded combat, lightning skills |
| `crowded_melee_perf_probe` | `tools/bot/scenarios/104_crowded_melee_perf_probe.json` | Crowded melee (v371+) |
| `dungeon_combat_perf_probe` | `tools/bot/scenarios/103_dungeon_combat_perf_probe.json` | Dungeon descent + combat render |

```bash
ARPG_PERF_DEBUG=1 make bot scenario=crowded_melee_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=dungeon_combat_perf_probe
```

### Benchmark scenarios (`make benchmark`)

```bash
make benchmark                                            # opens Godot, generates report
make benchmark BENCHMARK_OUT=docs/performance/reports/run.txt
```

Flow: server starts with `ARPG_PERF_DEBUG=1` → protocol bot records all benchmark scenarios (saves
replay manifest) → **Godot opens in visual mode** with the Performance status overlay active in the
top-right corner → after Godot closes, a perf report is generated from the server log.

If Godot is not on `PATH`, the script degrades to protocol-only and still generates the report.
Set `GODOT=/path/to/godot` to override the binary.

Artifacts saved under `.artifacts/benchmark-runs/<timestamp>/`:
- `server.log` — raw structured server output (all `backend_perf` lines)
- `bot.log` — bot stderr (scenario begin/done markers used to slice the report)
- `manifest.json` — replay manifest for the Godot visual playback
- `report.txt` — per-scenario perf summary

Report sections per scenario: tick budget (overruns, max overrun ms), simulation phase breakdown
(total/sim/ai/pathfind/combat/broadcast/persist — avg, p95, max), pathfinding (requests, cache hit %,
nodes visited), entity load (monsters moved, changes, events, clients).

| Scenario ID | File | Focus |
|-------------|------|-------|
| `sorcerer_multigroup_perf_probe` | `tools/bot/scenarios/105_sorcerer_multigroup_perf_probe.json` | Sorcerer vs 18 dungeon_mob + 12 dungeon_undead + 6 dungeon_wolf (real 3D models, ~360k primitives); flee/chase, 3-skill rotation — **coop+Godot observer** |
| `paladin_charge_loop_perf_probe` | `tools/bot/scenarios/106_paladin_charge_loop_perf_probe.json` | Paladin Holy Shield + 10 charge passes through same 36 real-model enemies (~398k primitives): push/stun fanout, quadruped/skeleton mix — **coop+Godot observer** |
| `sorcerer_dungeon_perf_probe` | `tools/bot/scenarios/107_sorcerer_dungeon_perf_probe.json` | Sorcerer in a generated D1 dungeon: approach packs, rotate 3 skills for 40s — server-side only (`benchmark_solo_session: true`; dungeon world requires solo session) |

**Adding a benchmark scenario**

1. Set `"ci_tier": "benchmark"` in the JSON — automatically discovered by `make benchmark`. No other registration needed.
2. Prefer a compact lab world with pinned seed for coop+Godot scenarios (e.g. `crowded_lightning_perf_probe`).
3. For multi-level dungeon worlds (`dungeon_depth_one_lab` etc.): also set `"benchmark_solo_session": true`. This runs the bot in solo mode (correct dungeon spawning) and skips the Godot observer; only server-side metrics are captured.
4. Add a row to the table above and [Changelog](#changelog).

**Adding an extended perf probe**

1. Set `"ci_tier": "extended"` and add a row to the extended table above.
2. Prefer compact **lab world** + pinned seed; avoid incidental navigation.
3. Document command in slice as-built and [Changelog](#changelog).
4. Register in `docs/progress/scenario-catalog.md` if pack membership changes.

---

## Investigation workflow (for agents)

1. Reproduce with `make play-debug` or a perf probe + `ARPG_PERF_DEBUG=1`.
2. Split bottleneck: `backend_perf` vs `[client-perf]` phase suffix.
3. If client-bound, rank `delta` sub-phases (`d_chg`, `d_ui`, `d_upsert_*`, …).
4. Cross-check Godot monitors on the same line (`draw_calls`, `objects`, `process_ms`).
5. Change one layer at a time; prefer data-driven tuning in `shared/rules/` or `shared/assets/`.
6. Verify: targeted test → probe scenario → `make client-unit` / `go test` as appropriate.
7. Record **tool changes** in [Changelog](#changelog); record **findings** in a separate investigation note if needed.

---

## Changelog

*Append a row when adding metrics, phases, probes, overlays, or config keys. Newest first.*

| Date | Area | What | Files / commands |
|------|------|------|------------------|
| 2026-06-29 | Client | Delta sub-phases (`d_prep` … `d_recon`), `d_upsert` / `d_upsert_m` / `d_upsert_player`; ranked phase output; delta UI sync gate + `client_perf` interval keys | `perf_phase_timer.gd`, `perf_debug_sampler.gd`, `main.gd`, `delta_ui_sync_gate.gd`, `main_config.v0.json` |
| 2026-06-29 | Client | Local player upsert change-detection helpers | `local_player_authoritative_sync.gd` |
| 2026-06-29 | Client | `delta_ui_sync_interval_ticks`, `delta_minimap_sync_interval_ticks` | `main_config.v0.json`, `main_config_loader.gd` |
| 2026-06-29 | Bot | `paladin_charge_loop_perf_probe` (ci_tier=benchmark) — Holy Shield + 10 single-segment charge passes, push/stun fanout, monster re-pathfind after knockback | `tools/bot/scenarios/106_paladin_charge_loop_perf_probe.json` |
| 2026-06-29 | Bot | `sorcerer_multigroup_perf_probe` (ci_tier=benchmark) — 5-skill sorcerer, flee/chase cycle, teleport escape, multi-group pathfind | `tools/bot/scenarios/105_sorcerer_multigroup_perf_probe.json` |
| 2026-06-29 | Bot | `benchmark` tier + `make benchmark` command with perf report generation | `tools/bot/ci_pack.py`, `tools/bot/run.py`, `tools/bot/benchmark_report.py`, `scripts/benchmark.sh`, `make/ci.mk` |
| 2026-06-29 | Bot | `crowded_melee_perf_probe` | `tools/bot/scenarios/104_crowded_melee_perf_probe.json` |
| 2026-06-29 | Server | Combat tick budget, collision cache, persist defer, overload guardrails | `combat_tick_budget.go`, `tick_collision_cache.go`, `persist_defer.go`, `tick_guardrails.go` |
| 2026-06-29 | Client | Presentation LOD, projectile cap, recon backpressure, delta coalesce | v373–v378 slices; see CODEMAP |
| 2026-06-29 | Client | Client phase breakdown (`net_poll`, `delta`, `entities`, `fog`) | v370 — `perf_phase_timer.gd`, `main.gd` |
| 2026-06-18 | Client + Server | `ARPG_PERF_DEBUG`, `[client-perf]`, `backend_perf` logs | v267 — `scripts/play.sh`, `perf_debug_sampler.gd`, `realtime/perf_debug.go` |
| 2026-06-18 | Wire + UI | `state_delta.performance` + settings overlay | v272 — `performance_status_formatter.gd` |
| 2026-06-18 | Bot | `crowded_lightning_perf_probe` lab + scenario | v268 — `93_crowded_lightning_perf_probe.json` |
| 2026-06-26 | Bot | `dungeon_combat_perf_probe` | v347 — `103_dungeon_combat_perf_probe.json` |

<!-- Template:
| YYYY-MM-DD | Client / Server / Bot / Shared | Short description | paths or make commands |
-->

---

## Related docs

| Doc | Topic |
|-----|-------|
| [`docs/as-built/v267_perf-debug-mode.md`](../as-built/v267_perf-debug-mode.md) | Original perf debug mode |
| [`docs/as-built/v370_client-perf-breakdown.md`](../as-built/v370_client-perf-breakdown.md) | First client phase buckets |
| [`docs/as-built/v272_performance-status-overlay.md`](../as-built/v272_performance-status-overlay.md) | In-game performance panel |
| [`docs/as-built/v371_crowded-fight-perf-probe.md`](../as-built/v371_crowded-fight-perf-probe.md) | Crowded melee probe |
| [`PROGRESS.md`](../../PROGRESS.md) | Current slice baseline |

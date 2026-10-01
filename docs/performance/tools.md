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

For a focused live client-bot run, set `BOT_CLIENT_LOG_DIR=.artifacts/client-perf-logs`
alongside `ARPG_PERF_DEBUG=1 make bot-client SCENARIO=<id> HEADLESS=0`. The runner retains
`<id>-client.log` and `server.log` in that directory. The client log contains only
render counters and bounded `[client-frame-batch]` microsecond intervals; the server
log contains only `backend_perf` counters. This opt-in does not change default logs.

For the v497 live dungeon fixture, run one scenario per log directory:

```bash
ARPG_PERF_DEBUG=1 BOT_CLIENT_LOG_DIR=.artifacts/dungeon-frame make bot-client SCENARIO=dungeon_frame_pacing_probe HEADLESS=0
python3 -m tools.bot.dungeon_frame_report --client-log .artifacts/dungeon-frame/dungeon_frame_pacing_probe-client.log --server-log .artifacts/dungeon-frame/server.log --out .artifacts/dungeon-frame/report.txt
```

The fixture pins the generated seed, level, stationary isometric camera, Balanced tier,
1920×1080 window, Forward+ renderer, 31 walls, and 24 live monsters. The report rejects
missing frame batches, dropped intervals, short traces, or a changed scene. It separates
true per-frame interval p50/p95/p99 from percentiles of one-second averages, and shows
process, draw, primitive, and server tick costs separately. Repeat A/B only when the
renderer host is uncontended; keep each pair of raw logs and report.

To sample the Performance tier at the same resolution, use a separate log directory
and set `BOT_CLIENT_RENDER_QUALITY=performance` on the run, then pass
`--quality performance` to the report. This override is explicit in the retained
command and checked against every steady client batch.

```bash
ARPG_PERF_DEBUG=1 BOT_CLIENT_RENDER_QUALITY=performance BOT_CLIENT_LOG_DIR=.artifacts/dungeon-frame-performance make bot-client SCENARIO=dungeon_frame_pacing_probe HEADLESS=0
python3 -m tools.bot.dungeon_frame_report --client-log .artifacts/dungeon-frame-performance/dungeon_frame_pacing_probe-client.log --server-log .artifacts/dungeon-frame-performance/server.log --quality performance --out .artifacts/dungeon-frame-performance/report.txt
```

### First-spawn frame trace (v495)

Run a selected live bot scenario with a fresh visual Godot observer process in
each trial:

```bash
BENCHMARK_SCENARIO=sorcerer_multigroup_perf_probe \
BENCHMARK_RUNS=10 BENCHMARK_FIRST_SPAWN=1 \
BENCHMARK_BASELINE_LABEL=provisional-pre-v494 make benchmark
```

The command fails when Godot is absent, the bot fails, a trace has no first
monster frame, or the scenario's monster count, renderer, or quality tier varies
between trials. `.artifacts/benchmark-runs/<timestamp>/` contains each bot and
client log, `first-spawn-summary.json`, and `report.txt`. The summary includes
fixture seed, world, host, Git revision, Godot version, renderer, quality tier,
monster count, raw trial timings, phase ranking, frame intervals after spawn,
steady-frame distribution, draw calls, resource/node counts, and static memory.

`ARPG_FIRST_SPAWN_TRACE=1` enables `[client-spawn-frame]` JSON records. The
optimized town scene also logs `[client-startup] town_dressing_ms=...` before
the first network snapshot; include that shifted startup cost in comparisons.
`first_spawn` marker is the first client frame with monsters. `process_wall_ms`
measures that frame's main `_process` work. Its visible frame interval appears
in the *next* record's `frame_interval_ms`, because the next frame start is the
first time the full interval can be observed. The interval leading into the
spawn frame is also reported so work is not hidden before the marker. Phase
buckets overlap (`net_poll`
contains snapshot/upsert work), so do not sum them. The one-second
`[client-perf]` series remains available for broader context but is not the
first-spawn acceptance metric.

“Cold” here means a new Godot process with no in-process resource cache. Asset
import occurs once before the trials; OS file and GPU driver caches can remain
warm. Compare before and after on the same host, renderer, quality tier, seed,
scenario, and trial method. Ten pre-v494 trials are provisional; v495 acceptance
starts with a fresh post-v494 baseline and ten paired after trials.

### How `make benchmark` works — live concurrent session

The benchmark runs the **protocol bot and Godot client simultaneously on the same live session**:

1. Server starts with `ARPG_PERF_DEBUG=1`.
2. For each benchmark scenario, the bot creates a listed co-op session and writes the session ID to a temp file.
3. Three seconds later — if the bot is still running — Godot launches and joins the same session as a
   second player (`ARPG_JOIN_SESSION_ID`). Vsync is **off** by default (`--disable-vsync`) so frame-time
   headroom is visible; `BENCHMARK_VSYNC=1 make benchmark` keeps the project's vsync.
4. The bot drives the scenario (combat, spells, movement). Godot renders what it sees in real-time.
5. When the bot finishes, Godot is closed. Each scenario's Godot stdout is kept as
   `<scenario>-client.log` and reported as its own CLIENT section.
6. The bot runs with `--skip-replay`, which skips both `/state` (server-side `replay.Reconstruct`) and
   `/replay` verification — the benchmark measures cost, not determinism. Correctness of the same probes
   is gated protocol-only in `make ci-full` (see tiers below).
7. `make benchmark` exits non-zero if any scenario's bot failed (the report is still written).

This captures real `[client-perf]` frame cost under actual server load — not a replay. The CLIENT
section of the report shows what a second player sees while the first player (bot) is actively
fighting. It reports the **first-spawn hitch** (first sample with entities on screen: snapshot apply +
model instantiation) separately, drops the next warmup samples, and shows true frame interval
p50/p95/p99 when frame batches are available. It labels `avg_frame_ms` as one-second averages
and reports `process_ms`, `draw_calls`, and `primitives` separately. FPS is secondary: the report prints the
observer's vsync mode (`vsync=` field of `[client-perf]`) and flags a cap when the frame-time floor
(p25) equals the median. On macOS the cap persists even with `--disable-vsync` (Metal and MoltenVK
both block windowed apps on the compositor's drawables — verified 2026-09-29, Godot 4.7.2, M4 Pro), so
on Mac read `process_ms` for headroom.

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
- **Inputs:** `--server-log` (Go structured JSON) + `--bot-log` (bot stderr with scenario markers) +
  repeatable `--scenario-client-log <scenario_id>=<path>` (Godot observer stdout per scenario)
- **Output:** per-scenario summary using `backend_perf` lines; slices by wall-clock timestamp from bot
  markers. Client statistics live in `tools/bot/benchmark_client_stats.py` (warmup split, hitch,
  nearest-rank percentiles); `--warmup-settle-samples` (default 2) controls how many post-hitch samples
  are dropped.

**Adding a backend metric**

1. Prefer **deterministic counters** in `perf_debug.go` for replay/bot use; wall-clock only in `realtime/`.
2. Extend `buildPerformanceStatus` + `logBackendPerf` + protocol `performanceStatusPayload` together.
3. Bump protocol schema if wire shape changes.
4. Update `performance_status_formatter.gd` if user-visible.
5. Add Go test in `realtime/perf_debug_test.go` or focused `*_test.go`.
6. Append [Changelog](#changelog).

---

## Bot scenarios & lab worlds (stress / regression)

### v499 live targeting measurement

`106_live_targeting_corrections` runs the real Godot client bot against the
`combat_control_lab` server session. It moves, chases the moving cave wraith,
waits for an authoritative attack event, then changes floor destinations. The
client bot uses the existing click bridge; this is a live socket session, not
an offline replay. Run both profiles on the same commit, Godot build, display
mode, and seed:

```bash
HEADLESS=1 make bot-visual scenario=106_live_targeting_corrections
ARPG_BOT_TRANSPORT_PROFILE=bounded_80_20 HEADLESS=1 make bot-visual scenario=106_live_targeting_corrections
make bot-visual scenario=106_live_targeting_corrections  # visible play-camera check
```

The `bounded_80_20` fixture schedules 80 ms base delay plus a fixed sequence of
offsets from -20 to +20 ms in **each** WebSocket direction. Client frame stalls
can make actual delivery later than the scheduled 60–100 ms window. Envelopes remain
ordered and their contents are unchanged. `local` (the default) adds no delay.
The profile is accepted only while `ARPG_BOT_CLIENT=1`; an unknown profile
fails the client bot before the scenario begins. The runner saves only
`[targeting-trace]` JSON records to `.artifacts/v499/targeting-trace.*`,
excluding account, token, session, entity, and target identifiers.

Each trace marks input receipt, command queue/transport dispatch, the first
visual-node move or attack response, authoritative acceptance/rejection,
reconciliation displacement, and process-frame stalls over 33.3 ms. Its
summary includes sample counts and p50/p95; missing responses remain missing
samples, never zero-latency samples. A 600 ms post-scenario observation window
allows final acks and visible movement to arrive. Frame `delta` is a client
process-frame proxy, so inspect visible play-camera footage alongside the
trace before claiming a rendered-frame or feel improvement.

Collect at least three runs per profile, alternate profile order when possible,
and compare matched p95 values with sample counts, frame stalls, Godot version,
commit, and host/display conditions. v493 measurements are provisional because
v496 attack timing and v497 frame pacing are still To do. Select a correction
only after both are integrated and a reproducible defect is identified.

Reproducible perf paths without manual dungeon walks. Two tiers:

### Test topology is part of the assertion

A performance scenario is valid only when it includes the component suspected of failing:

| Topology | What it proves | What it does not prove |
|----------|----------------|------------------------|
| Python protocol bot | Server simulation, persistence, fanout, generic WebSocket behavior | Godot buffers, polling, frame stalls, rendering, client reconnect behavior |
| Offline Godot replay (`make bot-visual` for protocol scenarios) | Deterministic presentation cost for recorded envelopes | Live socket backpressure, heartbeat, close behavior, concurrent server/client timing |
| Godot observer in a bot-driven co-op session | Live Godot rendering and receive path for a remote player | The authoritative local-player input/reconciliation path |
| Godot client-bot (`runner: godot_client`) | Live transport, input, reconciliation, delta application, and presentation | Multi-peer capacity unless the scenario explicitly joins peers |
| Interactive `make play-debug` | Actual player controls and machine-specific behavior | Repeatable CI regression proof by itself |

Rules for future performance/session-stability work:

1. Reproduce using the real failing topology before declaring mitigation complete. A Python socket
   cannot close a Godot transport bug, and offline replay cannot exercise live backpressure.
2. Assert lifetime failures, not only final state. A client that reconnects successfully may finish
   with `ws_open=true`; scenarios must assert a zero reconnect count when continuity is required.
3. Record close codes/reasons and server read/write errors. Generic “connection lost” logs cannot
   distinguish peer close, buffer exhaustion, network reset, or server write failure.
4. Correlate the exact disconnect window. Healthy backend `total_ms` beside high client `net_poll`,
   `delta`, or `process_ms` points to transport/presentation pressure, not simulation work.
5. Include burst shape, not just average load: largest envelope bytes, queued packet count, event and
   change counts, and resync snapshot cost matter more than mean throughput for heavy skills.
6. Validate test setup. Assert requested character class, skills, entity density, peer count, and
   actual cast acceptance; a rejected cast or wrong client type is not a performance proof.
7. Keep a short interactive soak after automation for platform-specific failures, but land a
   deterministic live-client regression scenario before closing the issue.

- **`ci_tier: extended`** — included in `make ci-full`, excluded from `make ci`. Manual run with `make bot scenario=<id>`.
- **`ci_tier: benchmark`** — excluded from `make ci`'s bot packs and from `--scenario all`. Launched with
  `make benchmark` (bot + visual Godot observer + report). Two gates keep them from rotting:
  `tools/bot/test_benchmark_scenarios.py` (in `make ci`: pinned `debug_progression` must sustain each
  probe's skill loop under the current mana rules — see `tools/bot/benchmark_mana_budget.py`) and a
  protocol-only `--scenario benchmark` run in `make ci-full` step 9 that also verifies `/state` + replay. Benchmark probes are
  exempt from the 15s scenario ceiling (CLAUDE.md rule 12); their declared `max_elapsed_s` is the budget.

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
BENCHMARK_VSYNC=1 make benchmark                          # observer keeps project vsync
```

Flow: see [How `make benchmark` works](#how-make-benchmark-works--live-concurrent-session).

If Godot is not on `PATH`, the script degrades to protocol-only and still generates the report.
Set `GODOT=/path/to/godot` to override the binary.

Artifacts saved under `.artifacts/benchmark-runs/<timestamp>/`:
- `server.log` — raw structured server output (all `backend_perf` lines)
- `bot.log` — bot stderr (scenario begin/done markers used to slice the report)
- `<scenario>-bot.log` / `<scenario>-client.log` — per-scenario bot stderr and Godot observer stdout
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

1. Set `"ci_tier": "benchmark"` in the JSON — automatically discovered by `make benchmark` and the
   `make ci-full` protocol-only gate. Pin `debug_progression.stats.magic` high enough for the skill loop;
   `test_benchmark_scenarios.py` derives the requirement from `shared/rules` and fails with the shortfall.
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
| 2026-09-29 | Bot + Client + CI | `make benchmark` repaired: EXIT-trap-in-subshell server kill fixed, Godot observer launched without `tee` (real pid, log always written), vsync off by default (`BENCHMARK_VSYNC`), `--skip-replay` also skips `/state`, non-zero exit on scenario failure; report gains per-scenario client sections with first-spawn hitch, warmup exclusion, frame/process percentiles, draw load; `[client-perf]` gains `vsync=`; mana-sustainability test + ci-full protocol-only gate | `scripts/benchmark.sh`, `scripts/ci.sh`, `tools/bot/benchmark_report.py`, `tools/bot/benchmark_client_stats.py`, `tools/bot/benchmark_mana_budget.py`, `tools/bot/test_benchmark_scenarios.py`, `client/scripts/perf_debug_sampler.gd` |
| 2026-07-08 | Client + Server + Bot | Live Godot WebSocket capacity/close diagnostics and zero-reconnect crowded Volley proof; documented topology-aware performance testing | v457; `ranger_volley_live_session_stability`; `scripts/bot_client_local.sh` |
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

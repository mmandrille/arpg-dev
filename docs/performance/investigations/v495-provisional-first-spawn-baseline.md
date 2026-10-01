# v495 provisional first-spawn baseline (before v494)

**Status:** Provisional pre-v494 diagnosis and matched A/B. v494 dungeon dressing and its prop/draw-call baseline are not integrated in this worktree. These data cannot serve as v495's final acceptance proof.

## Fixture and method

- Checkout: v493-based worktree, Git HEAD recorded in the raw summary.
- Live topology: `sorcerer_multigroup_perf_probe` bot and a real, windowed Godot observer joined to the same session. Seed: `sorcerer_multigroup_perf_probe_seed`; 36 monsters in every observer snapshot.
- Host: Apple M4 Pro, macOS 26.7.1, Godot 4.7.2 stable; `forward_plus`, balanced graphics tier, vsync disabled.
- Cold state: new Godot process per trial, with in-process resources empty. Godot import occurred once before the series; OS and GPU caches may remain warm.
- Raw evidence: `.artifacts/benchmark-runs/20260930T215313Z/first-spawn-summary.json`, its ten `*-client.log` frame traces, ten `*-bot.log` files, and `report.txt` in this worktree. All ten bot scenarios completed and all ten client traces passed the probe's validity checks.
- Command: `BENCHMARK_SCENARIO=sorcerer_multigroup_perf_probe BENCHMARK_RUNS=10 BENCHMARK_FIRST_SPAWN=1 BENCHMARK_BASELINE_LABEL=provisional-pre-v494 make benchmark`.

## Raw first-spawn samples

All timings are milliseconds. `process` is the measured main `_process` work on the first frame with monsters. `prior interval` leads into that frame; `spawn interval` is observed when the next frame starts. Draw calls are taken from that next record, after monsters appear.

| Trial | Server tick | Monsters | Process | Prior interval | Spawn interval | Draw calls |
|------:|------------:|---------:|--------:|---------------:|---------------:|-----------:|
| 01 | 56 | 36 | 639.9 | 16.8 | 697.0 | 519 |
| 02 | 50 | 36 | 380.5 | 24.0 | 427.9 | 532 |
| 03 | 46 | 36 | 381.0 | 99.3 | 429.6 | 520 |
| 04 | 50 | 36 | 593.9 | 16.6 | 651.8 | 519 |
| 05 | 50 | 36 | 526.7 | 16.7 | 581.9 | 532 |
| 06 | 49 | 36 | 521.5 | 16.2 | 570.1 | 511 |
| 07 | 46 | 36 | 512.3 | 223.2 | 573.5 | 535 |
| 08 | 49 | 36 | 542.8 | 16.6 | 597.6 | 532 |
| 09 | 49 | 36 | 573.7 | 16.4 | 630.2 | 532 |
| 10 | 47 | 36 | 445.6 | 155.5 | 490.4 | 530 |

Nearest-rank process p50/p95/max: **521.5/639.9/639.9 ms**. Spawn interval p50/p95/max: **573.5/697.0/697.0 ms**. With ten samples, nearest-rank p95 equals the maximum. The varying prior intervals and server ticks show observer startup/scheduling variation even with a fixed scenario and seed; retain those fields in the post-v494 comparison.

## Phase diagnosis and next gate

Median first-spawn snapshot phases: entity upsert **239 ms**, UI follow-up **158 ms**, snapshot setup and clear **110 ms**, and world build **9 ms**. Monster upsert contributes **192 ms** within entity upsert; phases overlap and must not be added together. The setup and clear bucket includes town ground dressing construction, rather than only entity removal. Focused temporary timers measured about **100–126 ms** for that town build, **90 ms** for repeated skill-panel redraws, and about **50 ms** for redundant monster material tinting. Some trials also have a second frame above 100 ms after spawn. Draw calls on the first rendered monster frame ranged from 511 to 535. Steady frame p95 across trials ranged from 16.1 to 21.8 ms; peak static memory was 178.8 to 182.1 MiB.

## Quiet-host matched A/B after the bounded fix

The owner authorized a provisional implementation before v494 integration. Both series used the same measurement code, fixture, seed, 36 monsters, host, renderer, quality tier, vsync setting, fresh Godot observer per trial, and successful bot completion. For the before series, only the production optimization files were temporarily restored to HEAD; the optimized files were then restored for the after series. No other renderer-heavy worktree was active during this A/B. The two series are independent cold-process runs, with OS/GPU caches potentially warm.

- Before: `.artifacts/benchmark-runs/20260930T223715Z/first-spawn-summary.json` and ten raw client/bot log pairs.
- After: `.artifacts/benchmark-runs/20260930T224134Z/first-spawn-summary.json` and ten raw client/bot log pairs.
- Command for each: `BENCHMARK_SCENARIO=sorcerer_multigroup_perf_probe BENCHMARK_RUNS=10 BENCHMARK_FIRST_SPAWN=1 BENCHMARK_BASELINE_LABEL=<label> make benchmark`.

| Run | Before process | After process | Before visible interval | After visible interval | After startup town build |
|---:|---:|---:|---:|---:|---:|
| 01 | 529.2 | 280.6 | 580.8 | 325.3 | 82.5 |
| 02 | 518.4 | 265.3 | 565.2 | 305.0 | 76.8 |
| 03 | 521.8 | 265.5 | 569.8 | 308.1 | 95.5 |
| 04 | 501.0 | 242.9 | 549.7 | 280.8 | 74.9 |
| 05 | 640.4 | 253.4 | 688.5 | 292.3 | 77.5 |
| 06 | 464.6 | 245.0 | 511.1 | 283.0 | 71.7 |
| 07 | 491.5 | 257.9 | 538.0 | 296.2 | 92.6 |
| 08 | 469.9 | 255.6 | 513.8 | 296.0 | 75.4 |
| 09 | 473.4 | 268.8 | 517.1 | 311.0 | 78.8 |
| 10 | 470.5 | 267.4 | 516.8 | 308.0 | 77.8 |

All times are ms. First-spawn `_process` p50/p95/max changed from **491.5/640.4/640.4** to **257.9/280.6/280.6** (p95 **56.2% lower**). Visible interval p50/p95/max changed from **538.0/688.5/688.5** to **296.2/325.3/325.3** (p95 **52.7% lower**). With ten samples, nearest-rank p95 equals max; the medians also improved substantially.

The fix shares source-tinted materials until a monster needs an individual hit or highlight reaction, bounds the tint cache to 64 entries, constructs town dressing once during initial scene setup and reuses it for same-level snapshots, and applies snapshot skill-panel state in one redraw. The moved town work took **71.7–95.5 ms** during startup in the after runs. It is a visible cost of the fix, though it occurs before the session's first monster frame; the corresponding before construction was inside the first snapshot. The first five intervals *after* the spawn interval reached at most **33.6 ms before** and **32.4 ms after**, so this probe shows no new multi-frame catch-up pause.

Guardrails from the same ten-run summaries: trial steady-frame p95 ranges were **12.6–14.3 ms before** and **11.4–13.1 ms after**. First rendered spawn-frame draw calls changed **511 → 505**. Maximum observed resource count changed **949 → 945**, maximum node count stayed **4995**, and maximum static memory changed **179.55 → 176.01 MiB**. These are observations on this fixture, not general game-wide guarantees.

After v494 lands, repeat a fresh ten-run v494 baseline and an integrated after series. The final gate must inspect any changed dominant phase and confirm the target, appearance, startup and transition costs, catch-up frames, steady pacing, resources, and memory on the combined build.

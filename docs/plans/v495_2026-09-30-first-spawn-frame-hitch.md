# v495 Plan — First-spawn frame hitch

**Status:** Complete; final combined `make ci` passed in 11m13s.
**Goal:** Measure the actual first monster frame in a live Godot session, then remove its dominant measured cost without changing authoritative gameplay or visual content.
**Architecture:** Extend the existing opt-in client performance sampler with per-frame trace records, retaining the existing one-second report. Run the existing `sorcerer_multigroup_perf_probe` bot and a fresh Godot observer process together for each trial. The measured pre-v494 costs led to a bounded material cache, a single reusable town dressing root built during initial scene setup, and one batched skills-panel redraw per snapshot. Repeat the same fixture after v494 lands before accepting the slice.
**Tech stack:** Godot/GDScript client, Bash benchmark runner, Python report and tests. No server, protocol, gameplay-rule, or asset contract changes expected.

## Baseline and shortcut decision

- v493 is the latest completed slice in `PROGRESS.md`; v494 room dressing is still To do. Its prop-count and draw-call baseline is a required dependency, so measurements on this checkout are provisional.
- **Borrow:** the in-repo KayKit manifests, `PerfDebugSampler`, `PerfPhaseTimer`, live benchmark topology, and `sorcerer_multigroup_perf_probe` seed. **Reject:** new assets, plugins, and speculative catalog-wide preloading.
- The existing one-second `[client-perf]` sample can miss the hitch and conflates initial snapshot, entity instantiation, and warmup. Use per-frame wall time and phase data for the acceptance metric. Define a cold run as a new Godot process with its in-process resource cache empty; record that OS/GPU caches may remain warm.
- The bot scenario must have its initial monster count and seed recorded on every run. An observer that fails to join or produces no spawn trace is a failed trial, not a low-latency sample.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/perf_phase_timer.gd`, `client/scripts/perf_debug_sampler.gd`, `client/scripts/main.gd` | Opt-in, per-frame trace at the real client update boundary, including spawn marker, phase timings, frame interval, process wall time, draw/resource counts. |
| Modify | `scripts/benchmark.sh` | Select one scenario, repeat it with a fresh observer process, require a live client for trace mode, and preserve each trial's logs and metadata. |
| Create | `tools/bot/first_spawn_report.py`, `tools/bot/test_first_spawn_report.py` | Validate trace completeness; summarize raw trial p50/p95/max, phase attribution, follow-on frames, steady state, resource and draw load. |
| Modify | `docs/performance/tools.md` | Document exact command, cold-run definition, metric meaning, and artifacts. |
| Modify | `client/scripts/model_tint.gd`, `client/scripts/model_reaction_controller.gd`, `client/scripts/town_dressing.gd`, `client/scripts/town_ground_detail.gd`, `client/scripts/skills_panel.gd` and focused tests | Address the measured material, town dressing, and UI redraw costs; retain no more than 64 tinted materials. |
| Later | `docs/as-built/v495_first-spawn-frame-hitch.md`, `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Close slice only after paired v494-based proof and final gates. |

## Maintenance ratchet

- `client/scripts/main.gd` is grandfathered at over 600 lines. Keep its trace hook and snapshot timing boundaries small; place detailed measurement logic in the sampler/timer files. Check the file-size ratchet after the final edits.
- Keep new Python files below 600 lines. Check all touched grandfathered files against `.maintainability/file-size-baseline.tsv` and run `make maintainability`.

## Task 1 — Capture the actual frame

- [x] Add opt-in per-frame records without changing the existing `[client-perf]` cadence or game behavior. Include frame index, tick, first-monster marker/count, previous frame interval, measured `_process` wall time, phase breakdown, node/resource/draw counts, and renderer/quality metadata in the run manifest.
- [x] Add focused Godot coverage for trace gating and spawn detection where feasible.

Verify: `make client-unit`; inspect one real-renderer trace for a spawn marker and consecutive frames.

## Task 2 — Repeatable live trial runner and report

- [x] Add selected-scenario repeat mode to the live benchmark while preserving the default full-suite command.
- [x] Start a fresh Godot observer in each trial; reject missing client, missing spawn marker, failed bot, or incomplete trace. Preserve one log pair per trial and a manifest with commit, host, Godot version, renderer, quality, seed, model count, and cold-process definition.
- [x] Summarize at least ten valid trial samples with nearest-rank p50/p95/max for first-spawn process wall time and frame interval, plus post-spawn catch-up frames, steady-frame p95, draw calls, nodes/resources, and ranked phases. Keep raw records.
- [x] Test parsing and incomplete-trial rejection with focused Python fixtures.

Verify: `python3 -m pytest tools/bot/test_first_spawn_report.py -q`; `bash -n scripts/benchmark.sh`; run one live trace trial.

## Task 3 — Provisional baseline on this checkout

- [x] Run at least ten valid cold-process trials against the current v493-based tree using `sorcerer_multigroup_perf_probe`. Save raw logs and report under `.artifacts/`; label the result provisional and record confounders.
- [x] Identify the largest measured phase/resource operation. Do not infer improvement from the historical 761 ms sample or from a one-second aggregate.

Verify: report shows trial count, identical seed/scenario, all spawn markers, and no failed trial.

## Task 4 — measured fix and v494 integration gate

- [x] After v494 landed in the integration baseline, reran ten cold-process trials with the same fixture and settings. Baseline p95 was 513.879 ms, so optimization remained warranted.
- [x] With owner approval, profile and implement a provisional pre-v494 fix: share tinted source materials until individual reactions need a private copy, cap the tint cache at 64, build town dressing in initial scene setup and reuse it for same-level snapshots, and batch skill-tree state updates into one redraw. Log the startup town cost and test visual/content behavior.
- [x] Run ten matched cold-process trials before and after on a quiet host. Provisional process p95 fell **640.4 → 280.6 ms (56.2%)** and visible interval p95 fell **688.5 → 325.3 ms (52.7%)**; raw samples and other guardrails are in the investigation note.
- [x] Repeated ten v494-based before/after trials after integration. First-spawn process p95 fell 513.879 → 279.275 ms (45.7%) and visible interval p95 fell 560.657 → 321.537 ms (42.7%), with lower max and no material guardrail regression. See the as-built report.

Verify: focused Godot tests; `make client-unit`; `make validate-assets`; `make bot-visual scenario=sorcerer_multigroup_perf_probe` (real-renderer visual check); inspect a first-group recording or frame trace.

## Task 5 — Lifecycle and final verification

- [x] Saved paired raw samples and summary in the v495 as-built; recorded fixture, renderer, tier, host, model count, and warm/cold state.
- [x] Update `PROGRESS.md` and `docs/progress/slice-lifecycle.md` once the batch is complete.
- [x] Run `make maintainability`, `make client-unit`, `make validate-assets`, and `make bot-visual scenario=sorcerer_multigroup_perf_probe` in this worktree. The `scenes` screenshot suite captured 9/9 images; inspect the town and monster PNGs. The coordinating task runs one combined `make ci` after integrating slices, per owner instruction.

## Deferred scope

The combined CI gate and lifecycle closeout were completed by the coordinating task. General draw-call batching belongs to v497.

Provisional run evidence: `docs/performance/investigations/v495-provisional-first-spawn-baseline.md` and raw `.artifacts/benchmark-runs/20260930T223715Z/` (before) and `20260930T224134Z/` (after) logs in this worktree.

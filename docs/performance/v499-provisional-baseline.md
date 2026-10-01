# v499 provisional live targeting baseline

Date: 2026-09-30 (America/Argentina/Cordoba). Source: detached `a2e8933c` plus the v499 measurement changes in this worktree. Host: Apple M4 Pro, Godot `4.7.2.stable.official.ed1daf0bf`, headless 1280×720 client bot, local Go server and per-worktree Postgres database. v496 and v497 are still **To do** in `PROGRESS.md`; these numbers are apparatus proof and must be recaptured after those slices integrate.

## Method

The live `106_live_targeting_corrections` scenario uses `combat_control_lab` and seed `live_targeting_499`. It drives the local hero through a floor move, monster chase, authoritative attack event, floor escape, and rapid floor retarget. The default profile adds no transport delay. `bounded_80_20` schedules outbound and inbound whole WebSocket envelopes after 60–100 ms each, using a fixed five-offset sequence and preserving order; client frame stalls can postpone actual delivery. Three fresh sessions per profile were run in alternating order. Each action has a 600 ms observation window at scenario end; a deliberately superseded floor click may have no visible response.

`[targeting-trace]` records only action kind, monotonic relative times, command kind, server tick, displacement, and frame duration. The runner retained those records under `.artifacts/v499/`; it did not retain account, token, session, target, or entity identifiers. The table lists each run's nearest-rank p95 in milliseconds. The trace files are local generated evidence, not committed source.

| Profile | Run | Trace file | Visible samples / actions | Input→visible p95 | Input→dispatch p95 | Input→ack p95 | Frames >33.3 ms |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: |
| local | 1 | `targeting-trace.h1KZaf` | 4 / 5 | 78.60 | 101.44 | 205.97 | 37 |
| bounded | 1 | `targeting-trace.ofWwB6` | 5 / 5 | 352.55 | 163.26 | 320.11 | 19 |
| local | 2 | `targeting-trace.m4g9ZL` | 5 / 5 | 101.48 | 80.18 | 139.76 | 4 |
| bounded | 2 | `targeting-trace.WvYKka` | 4 / 5 | 257.98 | 162.16 | 324.71 | 4 |
| local | 3 | `targeting-trace.vqAoDu` | 5 / 5 | 98.23 | 78.38 | 145.04 | 4 |
| bounded | 3 | `targeting-trace.yMeKuB` | 4 / 5 | 334.72 | 176.28 | 314.55 | 13 |

Across the three runs per profile, pooled nearest-rank p95 input→visible was **101.48 ms** (14 samples) locally and **352.55 ms** (13 samples) with delay; input→dispatch was **101.44 / 176.27 ms** (15 / 15 samples), and input→ack was **205.96 / 324.71 ms** (15 / 15 samples). These are comparisons between transport profiles, **not** a before/after improvement claim. Both profiles reported zero late attack dispatches and zero late local swings. Maximum recorded reconciliation displacement was 1.0 world unit in both profiles; the trace does not yet establish whether that is a visible snap. Process-frame stalls varied substantially across repeated headless runs (local 37/4/4, bounded 19/4/13), so frame pacing cannot be attributed to targeting from these traces alone.

## Decision

The scenario and deterministic delay fixture are usable. This baseline does not justify a targeting or smoothing change: v496/v497 are absent, the run is headless, and no stale-action or visible-snap defect was reproduced. Repeat the six-run comparison on their integrated baseline, inspect visible play-camera footage, and then choose one measured owner or stop the slice if no defect appears.

Focused checks: `make client-unit` passed, `make maintainability` passed, and the v461 `80_movement_visual_smoothing` and v464 `103_combat_input_flow_polish` live scenarios passed. The standard `make bot-visual` command was attempted but this new worktree's Python tooling bootstrap stalled on a package-registry read timeout. The successful scenario runs used `scripts/bot_client_local.sh` with the already installed Python environment from the primary checkout; the server, Godot client, and database were still started for this worktree. No `make ci`, replay check, or visible play-camera capture was used as proof here.

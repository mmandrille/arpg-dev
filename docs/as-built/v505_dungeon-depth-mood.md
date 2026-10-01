# v505 As-Built — Dungeon depth mood

Date: 2026-10-01
Status: Integrated; the final cache-aware Performance process-p95 gate and combined batch `make ci` pass. The earlier no-cache miss remains documented.

## What changed

- Added a schema-backed catalog with separate fog, torch-light, and flame palettes for the existing `shallow_cave`, `sundered_halls`, and `deep_vault` IDs. All tunable values stay in shared data.
- Passed the existing palette identity into the scene-lighting and torch presentation paths. Balanced uses per-depth render fog; Performance continues to disable render fog. Torch placement, count, light allocation, gameplay reveal radius, ambient/key-light settings, and town lighting remain unchanged.
- Cached the merged dungeon render context and torch config internally by palette ID, rebuilding only when the active palette changes. Cache values remain private and read-only to consumers; town/depth transitions and unknown-ID fallback remain covered.
- Kept unknown palette IDs on the existing base render/torch profiles. Updated the render ownership row in `docs/CODEMAP.md`.

## Focused verification

- `make validate-shared` — PASS (2,210 shared checks; CODEMAP validation passed).
- `godot --headless --path client --script res://tests/test_dungeon_depth_mood.gd` — PASS, 4 checks. Covers all existing palette IDs, unknown-ID fallback, render-fog tier/town behavior, and live torch color refresh on depth transition.
- `make client-unit` — PASS on the final source state.
- `make maintainability` — PASS (file-size ratchet, extraction-coupling ratchet, and progress-dashboard check).
- `git diff --check` — PASS.

## Real-camera visual evidence

Used existing live-player-camera scenarios at native 1920×1080 Forward+ on Balanced and Performance, with HUD and gameplay fog-of-war active:

- `make bot-visual scenario=dungeon_light_readability_shallow HEADLESS=0`
- `make bot-visual scenario=dungeon_light_readability_boss_balanced HEADLESS=0`
- `make bot-visual scenario=dungeon_light_readability_boss_performance HEADLESS=0`
- `make bot-visual scenario=dungeon_light_readability_deep HEADLESS=0`
- `make bot-visual scenario=dungeon_light_readability_engaged_enemy HEADLESS=0`

Baseline frames and metadata are in `.artifacts/v505-baseline/visual/`; candidate frames and metadata are in `.artifacts/v505-matched/visual/`. The inspected captures retain player/enemy separation, the depth-5 boss warning, visible ground loot, warm torch pools, and the gameplay fog boundary. This is real-renderer camera evidence for those fixtures; it does not prove behavior on other hardware. The depth-8 Performance camera fixture does not contain a visible local torch, so depth-specific torch application there is covered by the runtime unit test instead.

## Matched Performance comparison

The first base timing batch overlapped another slice's renderer work and is retained only as discarded evidence under `.artifacts/v505-baseline/perf-*`. The pre-cache matched Performance comparison below used three interleaved exact-base/candidate runs on an idle renderer slot. Every `dungeon_frame_pacing_probe` run passed. The candidate was run after removing slice-added debug-state fields; that source also passed `make client-unit`.

Fixture: macOS 26.7.1, Apple M4 Pro, Godot 4.7.2, Metal Forward+, 1920×1080, vsync 1, Performance tier, stationary isometric `generated_wall_lab`, seed `dungeon_levels_fast_247`, depth −1, 31 walls, 24 live monsters. Each run has 30 one-second client batches and 2,465–2,538 steady frame intervals.

| Metric (median across 3 runs) | Exact base | Candidate | Delta | Limit | Result |
|---|---:|---:|---:|---:|---|
| Frame-interval p50 | 11.95 ms | 11.88 ms | −0.07 ms | — | — |
| Frame-interval p95 | 13.49 ms | 13.89 ms | +0.40 ms | +0.67 ms | PASS |
| Process p50 | 14.58 ms | 14.51 ms | −0.07 ms | — | — |
| Process p95 | 16.24 ms | 17.18 ms | **+0.94 ms** | **+0.812 ms** (`max(0.5 ms, 5% of base)`) | **MISS by 0.128 ms** |
| Draw calls p95 | 68 | 68 | 0 | +10 calls | PASS |
| First-spawn process median | 277.78 ms | 297.05 ms | +19.27 ms | +27.78 ms (`max(20 ms, 10% of base)`) | PASS |

Per-run process-p95 values (base → candidate) are 16.24→17.18, 15.85→17.46, and 16.41→16.35 ms. This pre-cache comparison missed the fixed process-p95 budget by 0.128 ms.

The earlier matched Balanced set under `.artifacts/v505-isolated/` measured process-p95 medians of 16.95→17.25 ms (+0.30 ms; 0.8475 ms allowance), frame-interval p95 of 14.10→14.43 ms, and 108 draw calls on both sides. That set predates removal of the slice-added debug-state fields and is contextual only, not an exact-final-source Balanced timing claim.

Raw logs and reports for this pre-cache Performance comparison are retained in `.artifacts/v505-recheck/{base,candidate}-performance-{1,2,3}/`.

### Palette-cache Performance rerun

After adding the palette-ID caches, the exact-base/candidate Performance comparison was repeated as three interleaved pairs on the idle renderer slot. All six `dungeon_frame_pacing_probe` runs passed. Fixture: macOS 26.7.1, Apple M4 Pro, Godot 4.7.2, Metal Forward+, 1920×1080, vsync 1, Performance tier, stationary isometric `generated_wall_lab`, seed `dungeon_levels_fast_247`, depth −1, 31 walls, and 24 live monsters. Each run recorded 30 one-second client batches and 2,335–2,533 steady frame intervals.

| Metric (median across 3 runs) | Exact base | Candidate with cache | Delta | Fixed limit | Result |
|---|---:|---:|---:|---:|---|
| Frame-interval p95 | 13.51 ms | 13.95 ms | +0.44 ms | — | — |
| Process p95 | 16.75 ms | 16.92 ms | **+0.17 ms** | **+0.812 ms** | **PASS; 0.642 ms headroom** |
| Draw calls p95 | 68 | 68 | 0 | — | unchanged |

Per-run process-p95 values (base → candidate) are 16.75→16.92, 17.81→17.41, and 16.18→16.15 ms. The process-p95 change passes the unchanged 0.812 ms allowance on this pinned depth −1 fixture. This result does not establish a cross-hardware result or deep-floor cost. Raw logs and per-run reports are retained in `.artifacts/v505-cache-ab/{base,candidate}-performance-{1,2,3}/`.

## Security and scope limits

The loader reads only the fixed repository-relative path for the checked-in shared catalog and parses structured JSON validated by `make validate-shared`. No user-controlled path/content, network input, secrets, or new dependency is introduced. This does not create a new application trust boundary; no separate security finding was identified.

Combined `make ci` passed on the integrated batch in 11m41s on 2026-10-01. `make ci-full` was not run. No commit or push was made from the slice worktree.

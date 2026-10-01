# v495 As-Built — First-spawn frame hitch

- **Date:** 2026-09-30
- **Spec:** [`v495_spec-first-spawn-frame-hitch.md`](../specs/v495_spec-first-spawn-frame-hitch.md) · **Plan:** [`v495_2026-09-30-first-spawn-frame-hitch.md`](../plans/v495_2026-09-30-first-spawn-frame-hitch.md)
- **Scope:** first live monster batch in the Godot client. No server, protocol, rules, or visual-content changes.

## Implementation

The opt-in first-spawn trace records the actual `_process` wall time and next visible frame interval, spawn count, phase timings, consecutive frames, renderer, quality, draw calls, node/resource counts, and memory. The repeat runner requires a fresh Godot observer and valid spawn marker for each trial; the report rejects incomplete runs.

Profiling identified three client operations on the first snapshot. Model tinting now reuses at most 64 source materials and creates a private material only when an individual model reaction needs it. The town dressing root is built during initial scene setup and reused on same-level snapshots. The skills panel batches snapshot state into one redraw. The town prebuild moves **76.228–87.263 ms** (median **78.729 ms**) to initial scene setup; it is an explicit startup cost.

## Matched integration measurement

The baseline included v494 room dressing and the other integrated work up to this point, with only the v495 optimization disabled. Both sides used `sorcerer_multigroup_perf_probe`, seed `sorcerer_multigroup_perf_probe_seed`, `benchmark_mixed_arena`, 36 monsters, 43 entities, Godot **4.7.2 Forward+**, **Balanced**, Apple M4 Pro, and 10/10 valid fresh-process trials each. The host was reserved for this A/B, without competing render jobs. Fresh process clears in-process caches; OS and GPU caches may remain warm.

| Metric | Before | After | Change |
|---|---:|---:|---:|
| First-spawn `_process` wall p95 | 513.879 ms | 279.275 ms | −45.7% |
| Visible frame interval p95 | 560.657 ms | 321.537 ms | −42.7% |
| First-spawn `_process` maximum | 513.879 ms | 279.275 ms | −45.7% |
| Spawn draw calls p95 | 509 | 503 | −6 |
| Maximum resources | 953 | 949 | −4 |
| Maximum nodes | 4,994 | 4,994 | unchanged |
| Maximum static memory | 179.923 MiB | 176.398 MiB | −3.525 MiB |
| Median of per-run steady frame p95 | 13.525 ms | 13.589 ms | +0.064 ms |
| Largest later interval in the five-frame window | 33.237 ms | 38.273 ms | +5.036 ms |

The latter interval is one frame, with no multi-frame catch-up pause in the recorded window. The steady p95 difference is 0.5% and does not indicate a material regression in this fixture. Median measured phase costs fell from **100.175 to 2.239 ms** for snapshot clear, **149.245 to 82.352 ms** for snapshot UI, and **182.794 to 126.466 ms** for monster upsert. Phase times overlap and should not be summed.

Raw logs and reports: `.artifacts/benchmark-runs/20260930T225402Z/` (before) and `.artifacts/benchmark-runs/20260930T225933Z/` (after). The earlier v493-only investigation is in [`v495-provisional-first-spawn-baseline.md`](../performance/investigations/v495-provisional-first-spawn-baseline.md); its numbers are not used for the acceptance comparison.

## Verification and limits

Focused integrated town nature/dressing/ground detail, skills panel, kit monster, room dressing, and trace tests passed. The client scenario and 10+10 live trials exercised actual first spawn. The final combined `make ci` passed in 11m13s. The first monster frame remains about 322 ms at p95: the reduction meets the specified 30% target but does not eliminate the visible hitch. Subsequent optimization should profile the remaining monster model upsert and skills redraw cost on the full combined build.

# v500 town moving-camera comparison fixture

This fixture is used on the v497/v498 integrated baseline. The JSON route is outside
the normal bot scenario packs and runs only when passed by path.

## Fixed scene and run settings

- World: `dungeon_levels`, town level 0, seed `v500_town_render_route_20260930`.
- Route: six repetitions of the verified town floor positions `(4,8) → (19,12) →
  (11,16)` after a 30-tick entry settle, then 30 ticks at the final position.
  The isometric camera follows the moving player. The route deliberately avoids
  combat and UI panels so the scene composition stays comparable.
- Windowed Godot, Forward+ project renderer, `1280x720`, normal isometric zoom
  (`camera_presentations.v0.json` default size 12), same host/display and window
  placement, no extra capture or profiler overlay. `HEADLESS=0` is mandatory.
- Run Balanced and Performance separately with `BOT_CLIENT_RENDER_QUALITY=balanced`
  or `performance`; the runner applies that profile at launch.
- Keep vsync, display refresh, OS power mode, Godot version, and background load
  identical. Record them with each run; a capped frame-time floor makes
  `process_ms` the more useful headroom measure.

## Commands, after v497/v498 integration

On the **v498-integrated baseline before applying v500**, select Balanced and run:

```bash
ARPG_PERF_DEBUG=1 BOT_CLIENT_RENDER_QUALITY=balanced \
  BOT_CLIENT_LOG_DIR=.artifacts/v500-perf/baseline-balanced \
  BOT_STEP_DELAY=0 SCENARIO="$PWD/docs/performance/fixtures/v500_town_render_route.json" \
  HEADLESS=0 ./scripts/bot_client_local.sh
```

Repeat with `BOT_CLIENT_RENDER_QUALITY=performance` and `baseline-performance` as the log directory.
After applying v500 on the same host, repeat with `candidate-balanced` and
`candidate-performance`. Run each pair at least three times with fresh output
directories (for example, suffix `-1`, `-2`, `-3`); preserve every raw log and
its scenario PASS result. The runner saves
`v500_town_render_route-client.log` in the chosen directory. Only allowlisted
performance counters are retained; full session and account logs are discarded. Use the same route,
resolution, quality tier, and input delay for both revisions.

The existing client-stat parser can summarize any retained log:

```bash
python3 - <<'PY'
from pathlib import Path
from tools.bot.benchmark_report import parse_client_perf_lines
from tools.bot.benchmark_client_stats import render_client_block

root = Path(".artifacts/v500-perf")
for label in ("baseline-balanced", "candidate-balanced",
              "baseline-performance", "candidate-performance"):
    for path in sorted(root.glob(label + "*/v500_town_render_route-client.log")):
        samples = parse_client_perf_lines(path)
        print("\n".join(render_client_block(path.parent.name, samples)))
PY
```

For each tier, compare steady `avg_frame_ms` p50/p95/p99, `process_ms` p95,
`draw_calls` and `primitives` p50/p95, and the first-spawn sample separately.
Keep sample counts. A run that fails its route or has too few steady samples to
resolve p95 (fewer than 20) must be repeated with a longer fixture; do not
substitute a screenshot or a headless run. Compare against both the matched
pre-v500 town baseline and the actual permitted v497/v498 frame budget. The
current draft v497/v498 specs do not provide a numeric town threshold.

## Other acceptance measurements

This route plus `[client-perf]` provides moving-camera frame and draw-load data.
The current sampler does **not** report process memory or a precise town-entry
elapsed time. Collect those separately with the v497/v498 live-client method on
the same four configurations, including the larger v500 ground plane and forest
texture imports. Record the metric definition, raw samples, and host. Inspect
moving crowns over characters/targets and an extreme town corner in a separate
real-window visual pass; this route alone does not prove those views.

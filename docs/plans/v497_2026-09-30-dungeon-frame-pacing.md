# v497 Plan — Dungeon frame pacing

- **Status:** Complete; combined `make ci` passed in 10m33s.
- **Goal:** Reduce the measured steady-state render cost of a dense dungeon without changing geometry, per-wall fade/pick behavior, or the established visual quality.
- **Architecture:** Profile the live Godot renderer first, then optimize only the dominant pass. Preserve each authoritative wall rectangle as its own `StaticBody3D` and independently addressable fade/pick unit. Floors already use a `MultiMesh` per kit variant; any wall batching must stay inside a wall's identity boundary. Keep tuning, if needed, in schema-backed presentation data.
- **Tech stack:** Godot 4 Forward+ client, shared presentation JSON, Python benchmark/report tooling, Go server only as the unchanged live fixture host.

## Spec review and execution gate

The v497 spec passes the scope and ownership review. It has no expected protocol, gameplay-rule, or server change. Its 25% draw-call target is conditional on a **fresh post-v496 live profile showing walls as a dominant source**. If props or shadow passes dominate, revise the target and chosen intervention in this plan before implementation. The historical fourfold comparison is context, not a current baseline.

At planning time `PROGRESS.md` named v493 as latest complete and v494–v496 as To do. Those exploratory measurements remain provisional; the later combined v494–v496 A/B is recorded in the [as-built](../as-built/v497_dungeon-frame-pacing.md). The scheduled review/refactor gate is tracked separately from this render change.

The existing `make benchmark` dungeon probe uses `benchmark_solo_session: true` and skips the Godot observer. The current client report calculates p99 from one-second `avg_frame_ms` windows, not raw per-frame times. Both gaps must be closed before a matched A/B can satisfy the spec.

## Baseline and shortcut decision

- **Borrow:** Existing KayKit meshes and atlas, `DungeonKitFloor` MultiMesh, wall/fade code, `SceneLightingRig`, client performance sampler, and benchmark report.
- **Reject:** New art packs, renderer plugins, quality-tier downgrades, fewer monsters, or a global wall merge.
- **Adopt:** No outside dependency. Evaluate Godot's existing per-wall mesh/batching and shadow-caster controls only after profiling.

An exploratory **static** Forward+ probe on an Apple M4 Pro with Godot 4.7.2 used 58 wall rectangles at dungeon level -1 and Balanced quality. Median draw calls were 35 with all geometry, 24 with walls hidden, 21 with the floor hidden, and 9 with key-light shadows disabled; Performance also showed 9. These are non-additive visibility ablations of a synthetic scene, without v494 dressing, monsters, movement, or a live server. They suggest that blindly merging wall pieces is unlikely to be the first change. The probe log is in `.artifacts/v497-probe/render-probe.log` in this worktree; it is not a final benchmark artifact.

The original short `wall_floor_dungeon_rollout` run passed in a real Godot window but did not retain enough samples. The opt-in longer run below supplies a provisional live profile; the final A/B still needs the post-v496 scene and uncontended hardware.

## Current live profile and revised target (provisional)

An opt-in retained client log and temporary SceneTree profiler extended the existing `wall_floor_dungeon_rollout` scenario. On this checkout, the pinned `dungeon_levels_fast_247` level -1 scene had 31 wall rectangles, 24 live monsters, Balanced quality, Forward+, and 971 flat floor instances. In the same live run, visibility ablations gave these median draw counts: all visible **86**, walls hidden **78**, floor hidden **76**, key-light shadows disabled **65**. They are diagnostic, non-additive ablations; hiding walls or all shadows is not an implementation option. Walls contribute only 8/86 calls, so the spec's conditional 25% wall draw-call target is **not applicable to this measured fixture**.

**Revised v497 target for this path:** on the post-v496 dense live fixture, reduce steady rendered primitives by at least 30% through a visually neutral source-specific change, with no frame-time p95 regression and a maintained Performance-tier benefit. Report draw calls separately; the current one-call reduction is not a 25% draw-call win. If the combined v494–v496 profile changes the dominant cost, reselect the optimization and target before integration.

The bounded prototype disables shadow casting only for the flat KayKit floor slab by schema-backed variant data; raised decorated/weeds variants still cast shadows. It does not alter wall nodes, collision, navigation, fade, or pick identity. Current live counts are **86 → 85 draw calls** and **170,785 → 106,699 primitives** (−37.5%) with the same seed/wall/monster counts. This satisfies the **provisional primitive-count criterion only**. An interleaved A/B/A/B run gave frame p95 (ms) **30.52 / 29.48 / 29.65 / 30.55** for shadow on/off/on/off, so it does not establish a repeatable frame-time improvement. Other worktrees were rendering concurrently on this host; the coordinating task must repeat decisive matched measurements without competing render workloads before retaining or integrating the prototype. If uncontended frame p95 regresses, revert it.

In a static 900×620 Forward+ room capture, the shadow toggle changed 42 of 558,000 pixels with maximum channel delta 0.0235. This is narrow visual evidence, not a substitute for shallow/deep play-camera comparison after v494.

## File map

| Action | Path | Responsibility |
|---|---|---|
| Modify | `scripts/bot_client.sh`, `scripts/bot_client_local.sh`, `client/scripts/client_settings.gd`, `client/scripts/perf_debug_sampler.gd`, `tools/bot/benchmark_report.py`, `tools/bot/benchmark_client_stats.py` | Retain both live logs, pin the bot render setup, and attach bounded per-frame samples to the existing one-second counters. |
| Add | `tools/bot/dungeon_frame_report.py` | Reject drifted or incomplete fixture logs and report true frame, draw, primitive, process, and server tick percentiles. |
| Add or modify | `tools/bot/scenarios/client/` and the smallest relevant runner hook | Pin a generated dungeon seed, depth, entity load, camera route, tier, and run duration for live Godot A/B. |
| Conditional modify | `client/scripts/dungeon_kit_wall_builder.gd`, `client/scripts/wall_renderer.gd`, `client/scripts/dungeon_kit_floor.gd`, or a focused new helper | Change only the measured wall/floor cost while keeping each wall independently faded and picked. |
| Prototype modify | `shared/assets/dungeon_kit_presentation.v0.json` and schema, `client/scripts/dungeon_kit_floor.gd`, `client/tests/test_dungeon_kit.gd` | Allow only the flat slab variant to skip a redundant shadow pass; keep raised variants and wall identity unchanged. |
| Conditional modify | `client/scripts/scene_lighting_rig.gd`, `client/scripts/render_environment_presentation.gd`, `shared/assets/render_presentation.v0.json` and schema | A bounded shadow change only if the live profiler identifies redundant shadow work and captures show no visible loss. |
| Modify | `client/tests/test_dungeon_kit.gd`, `client/tests/test_wall_occlusion_fade.gd` | Cover two adjacent walls with independent fade/pick state, alignment, coverage, and teardown. |
| Modify | `docs/performance/tools.md`, `docs/CODEMAP.md`, `docs/as-built/v497_dungeon-frame-pacing.md`, `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Explain the probe and record the final matched evidence and lifecycle only after implementation. |

## Maintenance ratchet

`wall_renderer.gd` is 549 lines now, near the 600-line cap; keep it flat or extract a focused helper if the selected change needs more than a small hook. The other likely client/tool files are below 600 lines and absent from `.maintainability/file-size-baseline.tsv`. Check any new runner file against the baseline before touching it. Run `make maintainability` after editing.

## Task 1 — Establish a matched live-renderer fixture

Files: `tools/bot/scenarios/client/`, `scripts/benchmark.sh` or the smallest existing client-bot runner hook, `client/scripts/perf_debug_sampler.gd`, `tools/bot/benchmark_client_stats.py`.

- [ ] After the prerequisite slices land, pin one generated dense dungeon seed/depth and a repeatable live-client camera/combat route. Assert the actual wall count, live monsters, props, quality tier, renderer, and run duration; fail if the fixture silently changes.
- [ ] Keep raw per-frame timing samples or a bounded histogram separate from the existing one-second summary, and report p50/p95/p99 plus process p95, draw calls, primitives, and server tick p95. Exclude first-spawn and warmup explicitly, with counts and reasons.
- [ ] Record hardware, Godot/renderer version, resolution, display cap/vsync, catalog revision, seed, entity count, and camera route in the report. Use identical settings and at least three complete repetitions per A/B side; compare matched runs and retain raw logs.

Verification:

```bash
ARPG_PERF_DEBUG=1 BOT_CLIENT_LOG_DIR=.artifacts/dungeon-frame make bot-client SCENARIO=dungeon_frame_pacing_probe HEADLESS=0
python3 -m tools.bot.dungeon_frame_report --client-log .artifacts/dungeon-frame/dungeon_frame_pacing_probe-client.log --server-log .artifacts/dungeon-frame/server.log --out .artifacts/dungeon-frame/report.txt
```

Keep one scenario per retained log directory. The report fails if steady frame batches are absent, truncated, or no longer match the pinned scene.
The Balanced fixture uses 1920×1080, matching the client's Performance-tier window size;
`BOT_CLIENT_RENDER_QUALITY=performance` with `--quality performance` on the report measures
the second tier at the same resolution (see `docs/performance/tools.md`).

## Task 2 — Identify the dominant draw and frame cost

Files: probe/report tooling from Task 1, `docs/performance/tools.md`.

- [ ] Capture Balanced and Performance on the same scene. Use the Godot render profiler or controlled component visibility to split wall pieces, floor variants, v494 props, monsters/VFX, and key/torch shadow passes; state that visibility ablations are non-additive.
- [ ] Compare draw calls and primitives with frame p95/p99 and process p95. A draw-call decrease without a frame-time benefit still needs a visual and cost explanation.
- [ ] Update this plan with the measured bottleneck and chosen approach before changing renderer code. Retain the 25% target only if wall geometry dominates; otherwise set a source-specific measurable target consistent with the spec.

Verification: retained profiler capture and raw fixture report for each quality tier.

## Task 3 — Implement one bounded render change

Files: only the client/data files selected in Task 2, plus focused tests.

- [ ] If walls dominate, batch within each wall ID (or otherwise reduce its draw pass count) while preserving its original `StaticBody3D`, shape, transforms, material variation, and independently addressable mesh/fade state. Do not merge across wall IDs or rebuild full meshes per frame.
- [ ] If shadow work dominates, remove only demonstrably redundant casters/passes. Keep key-light shadow and wall/prop detail where the play-camera comparison shows it matters. Do not globally disable effects or change the default tier.
- [ ] Keep floor variant distribution and coverage unchanged; preserve a meaningful Performance-versus-Balanced cost gap.
- [ ] Add a two-wall test: fade A while B stays opaque and pickable, restore A, then fade B; confirm collision/pick toggles and visual bounds. Check town and dungeon transitions for stale state.

Verification:

```bash
make client-unit
make validate-shared
make maintainability
```

## Task 4 — Matched A/B and visual proof

Files: retained reports/captures, `docs/as-built/v497_dungeon-frame-pacing.md`.

- [ ] Run the same dense fixture, seed, entity count, camera route, hardware, renderer, resolution, and tier before and after; compare draw calls, primitives, true frame p50/p95/p99, process p95, and server tick p95 separately.
- [ ] At Balanced, meet the final plan's source-specific primitive target without frame p95 regression. At Performance, verify a meaningful cost reduction relative to Balanced.
- [ ] Capture shallow and deep play-camera frames in the real renderer before/after. Inspect wall geometry/detail, independent occlusion fade and picking, floor coverage, biome variation, shadows, and combat legibility. Run the relevant `scenes` screenshot suite and inspect its PNGs.
- [ ] Run focused client scenarios and give the owner these visual commands:

```bash
make bot-client SCENARIO=wall_occlusion_fade HEADLESS=0
make bot-client SCENARIO=wall_floor_dungeon_rollout HEADLESS=0
make bot-visual scenario=dungeon_combat_perf_probe
make regen-screenshots SUITE=scenes
```

The last bot-visual command is a visual combat regression; final frame pacing must be proved by the **live** fixture from Task 1.

## Task 5 — Close the slice after the gate

Files: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, `docs/CODEMAP.md`, `docs/performance/tools.md`, `docs/as-built/v497_dungeon-frame-pacing.md`.

- [x] Record the selected optimization, matched A/B sample counts, raw logs, captures, fixture limitations, and remaining GPU/headroom uncertainty.
- [x] In this worktree, stop after focused tests, visual scenarios, and matched measurements; report changed files and evidence to the coordinating task. The isolated session did not commit or transfer.
- [x] After integration on `main`, the coordinating task ran the final combined `make ci` and cleaned up worktrees. The isolated v497 session did not run `make ci`.

```bash
make ci
```

## Deferred scope

Monster-count or dungeon-layout reductions, authoritative collision/navigation changes, first-spawn loading, broad renderer replacement, and v498 lighting/readability design remain separate work.

## Integrated acceptance

On the combined v494–v496 scene, three valid interleaved runs per side/tier met the revised source-specific target: Balanced primitives fell 35.7% with no frame p95 regression, while Performance kept its cheaper shadow-off tier. Walls accounted for a small fraction of draw calls, so the conditional wall-batching target and two-wall identity test were not applicable. Paired shallow and deep real player-camera frames after the v498 torch lifecycle fix show no evident floor seam or lost prop shadow. The [as-built](../as-built/v497_dungeon-frame-pacing.md) has hardware, exact samples, excluded crash, images, and limits. The owner-requested single combined `make ci` remains the final gate; no separate CI run was made in this slice worktree.

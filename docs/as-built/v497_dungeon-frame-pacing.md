# v497 As-Built — Dungeon frame pacing

- **Date:** 2026-09-30
- **Spec:** [`v497_spec-dungeon-frame-pacing.md`](../specs/v497_spec-dungeon-frame-pacing.md) · **Plan:** [`v497_2026-09-30-dungeon-frame-pacing.md`](../plans/v497_2026-09-30-dungeon-frame-pacing.md)
- **Scope:** rendering cost of the flat KayKit dungeon floor variant; no geometry, wall identity, fade, collision, navigation, server rule, or quality-tier change.

## Measured owner and implementation

The live visibility probe found walls accounted for only about eight of 86 draw calls in its original scene, so the spec's conditional wall draw-call target did not apply. The flat floor slab was a redundant key-light shadow caster across hundreds of instances. `dungeon_kit_presentation.v0.json` now gives that one variant `cast_shadow: false`; the schema validates the field, and the floor builder uses it per variant. Raised decorated and weedy variants keep their shadows. No wall is merged or made unpickable.

The opt-in `dungeon_frame_pacing_probe` pins `generated_wall_lab`, seed `dungeon_levels_fast_247`, level −1, 31 walls, 24 live monsters, 2 interactables, stationary isometric camera, Forward+ Metal, 1920×1080, and vsync on. Its report requires a passing live client, unchanged fixture counters, at least 15 steady one-second batches, 300 complete frame intervals, and matching server samples. The retained client log contains only performance counters and the scenario sentinel; session identifiers and ordinary gameplay logs are omitted.

## Matched live A/B

On the Apple M4 Pro with Godot 4.7.2, three interleaved shadow-on/shadow-off runs per tier passed the report. The comparison used the combined v494–v496 code and the same catalog except this variant field. The v498 torch lifecycle fix was not yet applied; the later v498 A/B separately measures the final torch-present scene. Values below are medians of each run's reported p95, except draw calls, which were constant within the fixture. Each Balanced run retained 1,591–1,646 steady frame intervals; Performance runs retained 1,630–1,712.

| Tier / metric | Shadow on | Shadow off | Result |
|---|---:|---:|---|
| Balanced frame interval p95 | 19.24 ms | 19.15 ms | no regression |
| Balanced process p95 | 22.65 ms | 22.31 ms | no regression |
| Balanced primitives p95 | 179,737 | 115,651 | −35.7% |
| Balanced draw calls | 94 | 93 | −1 |
| Performance frame interval p95 | 18.95 ms | 18.76 ms | no regression |
| Performance process p95 | 22.06 ms | 22.07 ms | effectively unchanged |
| Performance primitives p95 | 90,300 | 90,302 | unchanged within dynamic scene noise |
| Performance draw calls | 68 | 68 | unchanged |

Performance already omits the relevant shadow pass, so the variant toggle has no expected material effect there. In the retained setting Performance remains cheaper than Balanced (68 versus 93 calls; about 90k versus 116k primitives). Server tick p95 varied between runs without a server code change and is reported in the raw files; it is not evidence of a renderer-caused backend change. One Performance shadow-on attempt ended in a Godot abort and was excluded before the three valid pairs; its crash log is retained under `before2-crash`.

Raw run logs and reports: `.artifacts/v497-matched/balanced/{before1,after1,before2,after2,before3,after3}/` and the equivalent `performance/` paths. These ignored artifacts remain in the integration worktree until transfer; the condensed acceptance numbers are recorded here. The static 900×620 Forward+ room comparison in the isolated worktree changed 42 of 558,000 pixels with maximum channel delta 0.0235.

Paired live player-camera captures at shallow level 1 and deep level 4 are retained under `assets/v497/` as `shallow-before/after.png` and `deep-before/after.png`, with fixture manifests. Each used the same camera position and scenario with the flat floor shadow field toggled; all four final visual scenarios passed. Direct inspection shows the visible tile joints, props, player, walls, and torch-lit deep floor remain legible and visually similar. The captures show no evident seam or lost prop shadow at this viewpoint. This is visual acceptance at two pinned viewpoints, not an exhaustive camera sweep.

## Verification and limits

Focused floor/kit unit tests, shared schema validation, asset validation, the Python report tests, and the live pinned fixture pass. The `scenes` screenshot suite on the isolated worktree produced 9/9 frames. The combined final `make ci` is deferred to the single integration gate requested by the owner. The observed reduction is in primitive work, not a large FPS gain; vsync and the CPU/overlay work limit the visible frame benefit on this host. A lower-end GPU has not been measured.

Visual replay: `make bot-client SCENARIO=dungeon_frame_pacing_probe HEADLESS=0`; inspect shallow/deep regression with `make regen-screenshots SUITE=scenes`.

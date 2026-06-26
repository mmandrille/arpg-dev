# v347 Spec - Dungeon Render Performance

Status: Complete
Date: 2026-06-26
Codename: dungeon-render-performance

## Purpose

Improve dungeon combat frame rate and perceived responsiveness on the Godot client by reducing
avoidable per-frame fog work and introducing a player-selectable graphics quality preset.

Investigation on a **real generated D1 dungeon** (`dungeon_levels`, 24+ live monsters, repeated
`ligthing` casts) measured **~53 FPS median / ~36 FPS floor** with GPU rendering enabled, while the
headless CPU path held **~71 FPS**. The gap points to **fullscreen fog + per-frame LOS shadow
rebuilds**, not mesh complexity. This slice (phase 1 of a three-part performance program) ships
measurable wins without changing server authority, protocol, or gameplay tuning.

Phase 1 delivers:

1. **Fog LOS shadow caching** — rebuild shadow polygons only when hero position, camera, light
   radius, or wall/occluder layout meaningfully changes; not on every `_process` tick.
2. **Graphics quality preset** — **Balanced** (current behavior) vs **Performance** (lower
   resolution target and throttled fog shadow refresh cadence), persisted in client settings.
3. **Dungeon combat benchmark scenario** — commit and document `dungeon_combat_perf_probe` for
   repeatable before/after perf sampling.

Follow-up slices (not v347): `forward_plus` renderer migration (v348), client presentation smoothing
between 10 Hz authoritative ticks (v349).

## Non-goals

- No server tick-rate change (live authority remains 10 Hz per `server/internal/realtime/protocol.go`).
- No gameplay tuning: monster population, skill damage, fog radius gameplay values, or combat balance.
- No protocol or `shared/protocol/` schema bump.
- No production metrics dashboard, trace backend, or always-on perf logging (v267 stays opt-in).
- No full settings screen expansion (remap, fullscreen, accessibility) — graphics quality preset
  only.
- No `forward_plus` / renderer migration in v347 (document evaluation matrix only if useful for v348).
- No client movement tick interpolation in v347 (defer to v349).
- No merge-gate CI pack promotion unless paired with budget-neutral demotion per CI pack policy.

## Acceptance Criteria

### Fog LOS shadow cache

- `FogOfWarOverlay` does **not** call full LOS shadow polygon rebuild on every frame while the hero,
  camera mode, light radius, viewport size, and wall/occluder layouts are unchanged within configured
  tolerances.
- Cache **invalidates and rebuilds** when any of these change:
  - wall layout (`set_wall_layout`)
  - extra occluder layout (`set_occluder_layout`)
  - light radius / progression (`set_light_radius`, `set_progression`)
  - camera perspective mode (`set_perspective_camera`)
  - hero world position beyond a data-driven move epsilon (reuse or extend
    `organic_edge.rotation_move_epsilon` pattern)
  - viewport size change large enough to affect projected shadow geometry
- Existing fog presentation remains correct:
  - LOS shadows for walls, tall obstacles, and door occluders still appear.
  - Organic edge, gloom/core shadow layering, and isometric vs chest-view behavior unchanged in
    semantics (visual parity within normal floating-point drift).
- Unit tests prove cache hit (no rebuild) across consecutive ticks with static state and cache miss
  on layout change and hero move past epsilon.
- `fog_of_war_overlay.gd` does not grow past its maintainability baseline without net extraction
  paydown in the same slice.

### Graphics quality preset

- Client settings expose **Balanced** and **Performance** graphics quality modes, persisted in
  `user://settings.json`.
- **Balanced** preserves current default window size (`2560×1440` for new installs) and current fog
  shadow refresh behavior after caching.
- **Performance** applies both:
  - window size target **1920×1080** (via existing `ClientSettings` size normalization), and
  - a data-driven cap on fog LOS shadow rebuild frequency (e.g. at most once per N frames or per
    configured minimum interval while invalidation conditions are not met).
- Settings panel shows the preset; changing it applies immediately without restart.
- `make validate-shared` passes if new fog presentation tuning keys are added.

### Benchmark scenario and perf targets

- `tools/bot/scenarios/103_dungeon_combat_perf_probe.json` is committed and listed in
  `docs/progress/scenario-catalog.md` as an **extended** dungeon combat perf probe.
- Scenario contract: descend to D1 on `dungeon_levels`, assert **≥10** live monsters, pull **≥2**
  `monster_aggro` events, cast `ligthing` **≥4** times with lightning damage observed.
- Protocol bot and visual replay both pass within scenario budget (`max_elapsed_s` ≤ 55).
- As-built records `ARPG_PERF_DEBUG=1` before/after samples on the implementation host using:

  ```bash
  ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_combat_perf_probe
  ```

- **Rendered combat-phase targets** (samples where `live_monsters ≥ 10` and `tick ≥ 100` on the
  probe host used for v347 sign-off):
  - median FPS **≥ 60** (baseline ~53),
  - floor FPS **≥ 45** (baseline ~36),
  - documented reduction in per-second shadow rebuild count vs pre-slice baseline.
- Perf targets are **semantic goals** for this slice sign-off, not hardcoded CI assertions (avoid
  machine-specific tuning locks).

### Regression and process

- Existing fog client bot scenarios remain green:
  - `67_fog_of_war_overlay`
  - `68_fog_los_shadow_mask`
  - `73_door_fog_toggle`
- `make client-unit`, `make maintainability`, and `make validate-shared` green.
- `crowded_lightning_perf_probe` remains green (lab vault regression guard).

## Scope and Likely Files

### Client

- `client/scripts/fog_of_war_overlay.gd` — shadow cache, invalidation, optional refresh throttle
- `client/scripts/hero_visibility_field.gd` — keep pure; only adjust if cache extraction needs helpers
- `client/scripts/fog_presentation_loader.gd` — load new perf tuning keys
- `client/scripts/client_settings.gd` — `graphics_quality` preset enum + apply path
- `client/scripts/settings_panel.gd` — preset UI control
- `client/tests/test_fog_of_war_overlay.gd` — cache hit/miss tests
- `client/tests/test_client_settings.gd` or focused new test — preset persistence/apply (if no existing
  coverage)

### Shared

- `shared/assets/fog_presentation.v0.json` — e.g. `shadow_rebuild_move_epsilon`,
  `shadow_rebuild_min_interval_frames` (names finalized in plan)
- `shared/assets/fog_presentation.v0.schema.json` — schema for new keys
- `tools/validate_shared.py` — semantic range guards for new tuning fields

### Bot / docs

- `tools/bot/scenarios/103_dungeon_combat_perf_probe.json` — benchmark scenario (may already exist
  in working tree)
- `docs/progress/scenario-catalog.md` — catalog entry + baseline note
- `docs/plans/v347_2026-06-26-dungeon-render-performance.md`
- `docs/as-built/v347_dungeon-render-performance.md`
- `docs/progress/slice-lifecycle.md`, `PROGRESS.md` — on `/finish` only

### Explicitly out of scope (v347)

- `client/project.godot` renderer method change
- `server/internal/**`
- `shared/protocol/**`

Asset/plugin decision: **reject** external fog plugins, shader packs, and Godot addons. **Borrow**
the in-repo `FogOfWarOverlay` shader and `HeroVisibilityField` LOS builders. **Reject** new imported
art; this is code-path and settings optimization only.

## Test and Bot Proof

```bash
make validate-shared
make client-unit
make maintainability
make bot scenario=dungeon_combat_perf_probe
ARPG_PERF_DEBUG=1 make bot scenario=dungeon_combat_perf_probe
ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_combat_perf_probe
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask
HEADLESS=1 make bot-visual scenario=73_door_fog_toggle
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
```

Manual visual verification (recommended during implementation):

```bash
ARPG_PERF_DEBUG=1 make play
# Settings → Graphics quality → Performance vs Balanced in D1 combat
```

## Baseline Evidence (pre-v347)

Probe host samples from `dungeon_combat_perf_probe` investigation (2026-06-26), combat phase
(`live_monsters ≥ 10`, `tick ≥ 100`):

| Mode | FPS median | FPS min | Notes |
|------|------------|---------|-------|
| Headless (CPU) | 71 | 49 | `draw_calls=0` |
| Rendered (GPU) | 53 | 36 | `draw_calls` avg ~139, `primitives` avg ~106k |

Primary suspects: per-frame `_update_shadows()` in `fog_of_war_overlay.gd`, fullscreen fog shader
fill, default 2560×1440 window, `gl_compatibility` renderer (renderer change deferred to v348).

## Open Questions and Risks

| # | Question | Plan default |
|---|----------|--------------|
| Q-1 | Exact Performance preset throttle: frame cap vs wall-clock interval? | Frame cap from `fog_presentation` data, overridable per quality preset in client settings mapping |
| Q-2 | Should Balanced mode also use shadow cache only (no throttle)? | **Yes** — cache is universal; throttle is Performance-only |
| Q-3 | Expose cache hit/miss counters in fog debug state for bot assertions? | **Yes** — stable debug keys for unit tests; optional bot assertion if cheap |
| Q-4 | Extract shadow cache to `fog_los_shadow_cache.gd` if overlay grows? | **Yes** if needed for maintainability ratchet |

Risks:

- **Visual regression:** cached shadows lag briefly behind fast hero movement if epsilon too large.
  Keep epsilon aligned with existing `organic_edge.rotation_move_epsilon` scale; add unit test for
  invalidation on move past threshold.
- **Stale shadows after door open:** door/occluder layout updates must invalidate cache (existing
  `set_occluder_layout` / wall layout paths).
- **Perf target portability:** FPS numbers vary by GPU; as-built captures host-specific before/after,
  CI proves correctness not FPS thresholds.
- **fog_of_war_overlay.gd size:** file is baselined at 624 lines — prefer extraction over growth.

## Sequenced Follow-up Program

| Slice | Codename | Scope |
|-------|----------|-------|
| v348 | `forward-plus-renderer` | Evaluate and migrate `client/project.godot` renderer; visual regression matrix |
| v349 | `movement-tick-smoothing` | Client presentation interpolation between 10 Hz authority updates; no protocol change |

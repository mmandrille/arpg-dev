# v347 Plan — Dungeon Render Performance

Status: Ready for implementation
Goal: Raise dungeon combat FPS by caching fog LOS shadow rebuilds and adding a Performance graphics preset, with a repeatable D1 combat benchmark scenario.
Architecture: Extract a focused `FogLosShadowCache` module that owns invalidation keys, cached polygon payloads, and rebuild counters. `FogOfWarOverlay` keeps shader/motion work per frame but delegates shadow geometry to the cache (rebuild only on layout/camera/radius/viewport/hero-move invalidation). **Balanced** uses cache only; **Performance** additionally applies 1920×1080 window sizing and a data-driven minimum rebuild interval from `fog_presentation.v0.json`. No server or protocol changes.
Tech stack: shared JSON (`fog_presentation`), Godot GDScript client, Python protocol bot, existing `ARPG_PERF_DEBUG` sampling.

## Baseline and shortcut decision

Builds on v253–v273 fog stack, v267 perf debug mode, and v264 organic fog silhouette. Pre-slice benchmark (`dungeon_combat_perf_probe`, 2026-06-26): rendered D1 combat **median ~53 FPS / floor ~36**; headless CPU **~71 FPS** — GPU/fog bound, not mesh count.

Asset/plugin decision: **reject** external fog plugins/shader packs/addons and new imported art. **Borrow** in-repo `FogOfWarOverlay`, `HeroVisibilityField`, `FogPresentationLoader`, and `ClientSettings` size normalization.

Reuse existing probe JSON `tools/bot/scenarios/103_dungeon_combat_perf_probe.json` (already validated in investigation).

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/assets/fog_presentation.v0.json` | `shadow_cache` tuning: move epsilon, viewport epsilon, performance min rebuild interval |
| Modify | `shared/assets/fog_presentation.v0.schema.json` | Schema for `shadow_cache` block |
| Modify | `tools/validate_fog_presentation.py` | Range guards for new keys |
| Add | `client/scripts/fog_los_shadow_cache.gd` | Cache state, invalidation, throttle, rebuild counters |
| Modify | `client/scripts/fog_presentation_loader.gd` | Accessors for `shadow_cache` keys |
| Modify | `client/scripts/fog_of_war_overlay.gd` | Delegate `_update_shadows` to cache; wire quality throttle |
| Modify | `client/scripts/client_settings.gd` | `graphics_quality` enum, persistence, Performance → 1080p apply |
| Modify | `client/scripts/settings_panel.gd` | Balanced/Performance toggle + signal |
| Modify | `client/scripts/main.gd` | Wire preset signal → settings + fog overlay throttle (minimal tentacles) |
| Modify | `client/tests/test_fog_of_war_overlay.gd` | Cache hit/miss + invalidation tests |
| Modify | `client/tests/test_client_bot.gd` | Graphics quality parse/normalize tests (settings section) |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Optional fog cache debug int assertions (if bot uses them) |
| Add/Modify | `tools/bot/scenarios/103_dungeon_combat_perf_probe.json` | Commit benchmark scenario (`ci_tier: extended`) |
| Modify | `docs/progress/scenario-catalog.md` | Catalog entry + pre-v347 baseline note |
| Modify | `docs/specs/v347_spec-dungeon-render-performance.md` | Mark complete on `/finish` |
| Add | `docs/as-built/v347_dungeon-render-performance.md` | Before/after perf samples + shipped behavior |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines; grandfathered files stay under baseline.

Hotspot / over-limit files touched:

- [x] `client/scripts/fog_of_war_overlay.gd` — baseline **624**; must **not grow** net — extract cache to new file
- [x] `client/scripts/main.gd` — grandfathered; only narrow wiring (preset → fog throttle), no new domains
- [ ] `client/scripts/bot_assertion_handlers.gd` — only if fog debug keys added; keep delta small
- [ ] `tools/validate_shared.py` — no change expected (delegates to `validate_fog_presentation.py`)
- [ ] Other over-limit files from baseline: **none planned**

Decision:

- [x] **Extract** `fog_los_shadow_cache.gd` as importable `class_name` with direct unit tests (extraction independence).
- [x] Do not add renderer migration or `main.gd` coordinator growth beyond preset wiring.

Verification:

```bash
make maintainability
```

## Task 1 — Shared fog cache tuning contract

Files:

- Modify: `shared/assets/fog_presentation.v0.json`
- Modify: `shared/assets/fog_presentation.v0.schema.json`
- Modify: `tools/validate_fog_presentation.py`
- Modify: `client/scripts/fog_presentation_loader.gd`

- [x] Step 1.1: Add `shadow_cache` object with:
  - `move_epsilon` (default align with `organic_edge.rotation_move_epsilon`, 0.006)
  - `viewport_size_epsilon_px` (e.g. 1.0)
  - `performance_min_rebuild_interval_frames` (e.g. 3 — tune in implementation, not locked in tests)
- [x] Step 1.2: Extend schema `additionalProperties` / nested required fields; add semantic range guards in `validate_fog_presentation.py`.
- [x] Step 1.3: Add `FogPresentationLoaderScript.shadow_cache()` accessor + defaults merge in loader.

```bash
make validate-shared
.venv/bin/pytest tools/test_validate_fog_presentation.py -q
```

## Task 2 — Fog LOS shadow cache module

Files:

- Add: `client/scripts/fog_los_shadow_cache.gd`
- Modify: `client/scripts/fog_of_war_overlay.gd`

- [x] Step 2.1: Implement `FogLosShadowCache` (`class_name`, `extends RefCounted`) with:
  - `invalidate()` on layout/camera/radius/mode changes
  - `should_rebuild(hero_world, viewport_size, perspective, throttle_frames) -> bool`
  - `rebuild(...)` calling `HeroVisibilityField.build_shadow_polygons` and storing polygon payload
  - `rebuild_count`, `cache_hits`, `last_rebuild_reason` for debug/tests
- [x] Step 2.2: Replace direct per-frame rebuild in `_update_shader` → `_update_shadows` path with cache consult; still sync `Polygon2D` nodes from cached payload each frame (cheap) but skip geometry rebuild on cache hit.
- [x] Step 2.3: Extend `get_debug_state()` with `shadow_cache_hits`, `shadow_rebuild_count`, `shadow_cache_valid` (names stable for tests).
- [x] Step 2.4: Ensure invalidation fires from `set_wall_layout`, `set_occluder_layout`, `set_light_radius`, `set_progression`, `set_perspective_camera`, and hero move past `move_epsilon`.

```bash
make client-unit
```

## Task 3 — Fog cache unit tests

Files:

- Modify: `client/tests/test_fog_of_war_overlay.gd`
- Add: `client/tests/test_fog_los_shadow_cache.gd` (if overlay tests would grow too much)

- [x] Step 3.1: Direct-import test for `FogLosShadowCache`: two consecutive rebuild calls with static inputs → second is cache hit (`rebuild_count == 1`).
- [x] Step 3.2: Layout change → cache miss and `rebuild_count` increments.
- [x] Step 3.3: Hero move within epsilon → hit; move beyond epsilon → miss.
- [x] Step 3.4: Performance throttle: with `min_rebuild_interval_frames = N`, forced invalidation still respects interval unless hard-invalidate (layout change always rebuilds immediately).
- [x] Step 3.5: Register new test in `client/scripts/client_smoke.sh` if new file added.

```bash
make client-unit
```

## Task 4 — Graphics quality preset

Files:

- Modify: `client/scripts/client_settings.gd`
- Modify: `client/scripts/settings_panel.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_client_bot.gd`

- [x] Step 4.1: Add `GRAPHICS_QUALITY_BALANCED` / `GRAPHICS_QUALITY_PERFORMANCE` constants, default Balanced, persist `graphics_quality` in `user://settings.json`.
- [x] Step 4.2: `apply_graphics_quality()`:
  - **Balanced:** keep user-selected `window_size` (default 2560×1440 for new installs)
  - **Performance:** set window to 1920×1080 via `normalize_size` / `_fit_size_to_screen`
- [x] Step 4.3: Settings panel: labeled option group (reuse camera-mode button pattern), signal `graphics_quality_selected(mode)`.
- [x] Step 4.4: `main.gd`: on preset change, call `client_settings.apply()` and `fog_overlay.set_performance_throttle(enabled)` (or pass interval frames from loader when Performance).
- [x] Step 4.5: Unit tests in `test_client_bot.gd` for parse/normalize/default of `graphics_quality` (mirror camera_mode tests).

```bash
make client-unit
```

## Task 5 — Bot scenarios and catalog

Files:

- Add/Modify: `tools/bot/scenarios/103_dungeon_combat_perf_probe.json`
- Modify: `docs/progress/scenario-catalog.md`

- [x] Step 5.1: Commit `103_dungeon_combat_perf_probe.json` with `ci_tier: extended` (no pack promotion).
- [x] Step 5.2: Catalog entry describing: D1 descent, ≥10 monsters, ≥2 aggro, 4× `ligthing`, perf sampling command.
- [x] Step 5.3: Document pre-v347 baseline FPS table in catalog footnote (median 53 / floor 36 rendered).

```bash
make bot scenario=dungeon_combat_perf_probe
ARPG_PERF_DEBUG=1 make bot scenario=dungeon_combat_perf_probe
```

## Task 6 — Regression bot / visual proof

Files:

- Modify: `client/scripts/bot_assertion_handlers.gd` (only if adding fog cache debug assertions to client scenarios)

- [x] Step 6.1: Re-run existing fog client scenarios unchanged (organic edge + LOS must still pass).
- [x] Step 6.2: Run rendered perf probe and capture before/after `[client-perf]` lines for as-built.

```bash
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask
HEADLESS=1 make bot-visual scenario=73_door_fog_toggle
ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_combat_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
```

Manual (recommended for sign-off):

```bash
ARPG_PERF_DEBUG=1 make play
# Settings → Performance vs Balanced in D1 combat; compare status FPS
```

### Perf sign-off criteria (as-built, not CI-locked)

Combat-phase samples (`live_monsters ≥ 10`, `tick ≥ 100`) on implementation host:

| Metric | Pre-v347 | Target |
|--------|----------|--------|
| FPS median | ~53 | ≥ 60 |
| FPS floor | ~36 | ≥ 45 |
| Shadow rebuilds/sec | ~60 (every frame) | materially lower; document `shadow_rebuild_count` delta |

## Task 7 — Lifecycle docs

Files:

- Modify: `docs/specs/v347_spec-dungeon-render-performance.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v347_dungeon-render-performance.md`
- Modify: `PROGRESS.md`

- [x] Step 7.1: As-built with perf table, cache behavior summary, and verification commands.
- [x] Step 7.2: Lifecycle row + PROGRESS current status (on `/finish`).

## Final verification

- [x] `make validate-shared`
- [x] `make client-unit`
- [x] `make maintainability`
- [x] `make bot scenario=dungeon_combat_perf_probe`
- [x] `ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_combat_perf_probe`
- [x] `HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay`
- [x] `HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask`
- [x] `HEADLESS=1 make bot-visual scenario=73_door_fog_toggle`
- [ ] `make ci` (final gate on `/finish` — failed 2026-06-26 on pre-existing `TestGeneratedDungeonMonsterRarityGolden`, `boss_floor_gate`, headless equip smoke; unrelated to v347 client changes)

## Deferred (explicit)

| Item | Slice |
|------|-------|
| `forward_plus` renderer migration | v348 |
| Client movement smoothing between 10 Hz ticks | v349 |
| Merge-gate CI pack promotion for `dungeon_combat_perf_probe` | Only with budget-neutral demotion |
| Fullscreen / controls remapping / accessibility settings | PROGRESS deferred backlog |

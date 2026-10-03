# v533 Plan — Camera zoom-out

- **Spec:** [`v533_spec-camera-zoom-out.md`](../specs/v533_spec-camera-zoom-out.md) · **Baseline:** `dab60eb5`, standalone, no prerequisites, no sibling overlap (touches only `camera_presentations.v0.json`).
- **Review gate:** pass. Data-only change; no determinism, protocol, golden, replay or ratchet impact; asset/plugin decision: adopt the existing camera catalog, reject any new plugin or asset.

## Tasks

- [x] **1. Baseline.** Capture `town-play` plaza and vendor at the old zoom; run the pinned probe with `ARPG_PERF_DEBUG=1` and keep the `[client-perf]` counters.
- [x] **2. Test first.** Add data-derived bounds assertions to `test_camera_mode_settings.gd` (they hold for old and new data by design).
- [x] **3. Change data.** `zoom_default` 12.0 → 15.0, `zoom_max` 20.0 → 25.0.
- [x] **4. Visual and cost proof.** Same captures and counters at the new zoom.
- [x] **5. Focused run.** `test_camera_mode_settings`, `test_combat_vfx`, `test_fog_of_war_overlay`, `make validate-shared`, `git diff --check`.
- [x] **6. As-built and trackers.**

## Final gate

Focused verification only. Combined `make ci` runs once after v532–v537.

# v505 Plan — Dungeon depth mood

- **Status:** Complete; final cache-aware Performance process-p95 gate and combined batch `make ci` passed. Earlier pre-cache miss remains documented.
- **Spec:** [v505 dungeon depth mood](../specs/v505_spec-dungeon-depth-mood.md)
- **Goal:** Extend the existing three dungeon palettes with data-driven atmospheric fog and torch/flame accents while preserving v498 readability, gameplay fog-of-war, quality-tier behavior, and measured render cost.
- **Base commit:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`.
- **Date:** 2026-10-01.
- **Prerequisites:** v498 is present at this base. No other accepted slice is a code prerequisite; v505 can implement independently. Coordinate captures against the latest integrated state if v504/v508 changes the visible loot subjects before this slice is integrated.

## Architecture and boundaries

The base already maps level depth to a stable palette ID in `GroundWallFactory` and `DungeonDepthLighting`. Keep that selection as the one source of depth-band identity; do not repeat numeric depth cutoffs in lighting or torch code. Add a focused `shared/assets/dungeon_depth_mood.v0.json` catalog keyed by `shallow_cave`, `sundered_halls`, and `deep_vault`, with a matching schema. Each entry contains bounded render-fog parameters and torch/flame color/energy overrides. Continue to use `dungeon_generation.v0.json` for wall/floor colors and existing directional/ambient light values.

Apply only the selected fog override to the existing dungeon render context, and only when the active graphics tier enables render fog. Apply the selected torch overrides through the existing torch presentation path; preserve torch placement, count, fog reveal radius, and performance-tier rules. Do not modify gameplay fog-of-war state or authoritative visibility. Keep v498's ambient suppression, general exposure, contrast, and torch-pool tuning unchanged unless a paired live capture demonstrates a specific readability issue.

**Asset/plugin decision:** borrow the in-repo palette definitions, KayKit torches, and current render pipeline. Reject external assets, shaders, post-processing packages, and plugins.

**Security review:** the Meli security skill was consulted for the local catalog loader. It reads a fixed repository-relative, schema-validated asset; no user-controlled path/data, network, authorization, user-content rendering, or secret-handling surface is added. No separate security finding was identified.

## File map and ownership

| Action | Paths | Notes |
|---|---|---|
| Add | `shared/assets/dungeon_depth_mood.v0.json`, `shared/assets/dungeon_depth_mood.v0.schema.json` | Per-palette render fog and torch/flame mood; numeric bounds and hex-color format in schema. No duplicated depth boundaries. |
| Modify | `client/scripts/dungeon_depth_lighting.gd` | Return the selected palette ID with the existing key/ambient profile. |
| Add | `client/scripts/dungeon_depth_mood_loader.gd` | Resolve the trusted catalog by palette ID; document the safe base-profile fallback. |
| Modify | `client/scripts/scene_lighting_rig.gd`, `client/scripts/render_environment_presentation.gd` | Merge only the selected fog settings into the existing dungeon context; retain all current tier gates. Avoid changes to `main.gd`. |
| Modify | `client/scripts/dungeon_torch_presentation_loader.gd`, `client/scripts/dungeon_torch_lights.gd` | Apply selected torch accents without changing mount placement, count, reveal radius, or light allocation. |
| Add/modify | `client/tests/test_dungeon_depth_mood.gd`; existing `test_render_environment_presentation.gd`, `test_dungeon_depth_lighting.gd`, and `test_dungeon_torch_placement.gd` only as needed | Derive expectations from the catalog; prove palette coverage, selected runtime values, unknown ID behavior, and tier gating. |
| Reuse/add only if needed | `tools/bot/scenarios/client/dungeon_light_readability_*.json` | Reuse existing depth-1 loot, depth-5 boss, and depth-8 routes; add a new route only for a missing live-camera subject or tier. Keep it extended unless it adds unique merge-blocking coverage. |
| Update | `docs/CODEMAP.md`, `docs/as-built/v505_dungeon-depth-mood.md` | Add catalog/loader ownership to Render baseline row; write as-built proof after execution. |
| Preserve as ignored evidence | `.artifacts/v505-baseline/`, `.artifacts/v505-matched/`, `.artifacts/bot-captures/` if produced | Record exact retained paths and metadata in handoff; never claim screenshots prove performance. |

Keep new or modified client files below the maintainability ratchet. Do not grow `main.gd` or the already-over-limit `fog_of_war_overlay.gd`; extract a focused helper if the necessary logic cannot fit their current ownership cleanly.

## Ordered tasks

### 1. Capture the starting state and lock the measurement budget

- [x] Record starting SHA, clean/dirty state, Godot version, OS/GPU, renderer, viewport, vsync, selected quality tier, and current render catalog values.
- [x] Capture the base scenes at native resolution before tuning: depth-1 ground loot, depth-5 boss warning, and depth-8 vault, in Balanced and Performance. Keep the HUD and gameplay fog-of-war active. Save frame metadata under `.artifacts/v505-baseline/`.
- [x] Run at least three interleaved base control samples per tested quality tier on the pinned frame-pacing topology. Retain raw client/server reports and first-spawn traces. The original overlapped set is preserved as discarded; the isolated exact-base/candidate set is in `.artifacts/v505-isolated/`, with a final-source Performance recheck in `.artifacts/v505-recheck/`. See the as-built note for the fixed-budget result.
- [x] Confirm existing bot routes render the expected depths/subjects from the live player camera. If any `make bot-visual` wrapper does not run with a window, use `make bot-client ... HEADLESS=0` for the capture and record the exact command used.

```bash
make bot-visual scenario=dungeon_light_readability_shallow HEADLESS=0
make bot-visual scenario=dungeon_light_readability_boss_balanced HEADLESS=0
make bot-visual scenario=dungeon_light_readability_boss_performance HEADLESS=0
make bot-visual scenario=dungeon_light_readability_deep HEADLESS=0
make bot-visual scenario=dungeon_light_readability_engaged_enemy HEADLESS=0
ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_frame_pacing_probe HEADLESS=0
```

Initial controls were captured before implementation but overlapped another slice's timed runs on this host. Preserve the raw six-run set as contention-discarded evidence; rerun the matched baseline/candidate A/B sequence after implementation with the other slice paused. Earlier camera captures remain archived for visual comparison.

### 2. Add and validate the depth-mood catalog

Files: the new shared asset JSON and schema, plus the focused loader.

- [x] Define one profile for every existing palette ID. Keep shallow cave warm, retain the established central-halls transition, and preserve the cooler deep-vault direction; choose final fog/torch values from paired live captures.
- [x] Put all tuning values in schema-backed shared data. Bound fog density/energy and torch energy; validate colors as six-digit hex. Do not put numeric depth cutoffs or tunable values in GDScript.
- [x] Make the focused loader return a deep copy and a documented safe base profile for an unknown ID. The new test must fail if an existing biome palette has no mood entry, so the runtime fallback does not hide catalog drift.
- [x] Do not add a Python cross-catalog rule to `tools/validate_shared.py`; the focused test/schema enforce the catalog contract.

```bash
make validate-shared
godot --headless --path client --script res://tests/test_dungeon_depth_mood.gd
```

### 3. Route the palette identity to existing render/torch owners

Files: `dungeon_depth_lighting.gd`, the new loader, `scene_lighting_rig.gd`, `render_environment_presentation.gd`, and existing torch presentation scripts.

- [x] Add the already-resolved palette ID to the lighting result without changing town/night resolution or current directional/ambient behavior.
- [x] Apply only the selected `fog` values to the dungeon render context. The existing `quality_tiers` remain authoritative: Performance still gates render fog off; Balanced still permits it.
- [x] Resolve the same palette ID for dungeon torches and apply only configured color/energy accents. Preserve the existing number of light nodes, cap, shadow policy, placement, and gameplay fog reveal behavior.
- [x] Prove that a level transition updates the world environment and torch presentation to the new profile and that a return/town transition does not leak dungeon mood settings into town.
- [x] Update `docs/CODEMAP.md` for new file ownership.

```bash
godot --headless --path client --script res://tests/test_dungeon_depth_mood.gd
godot --headless --path client --script res://tests/test_render_environment_presentation.gd
godot --headless --path client --script res://tests/test_dungeon_depth_lighting.gd
godot --headless --path client --script res://tests/test_dungeon_torch_placement.gd
make validate-shared
make maintainability
```

### 4. Tune and verify in the real renderer

Files: only the new mood catalog/schema unless a focused runtime defect blocks the spec.

- [x] Iterate one effect at a time, first render fog, then torch color/energy. Retain paired frames for accepted and rejected trials. Tune existing ambient palette values only when the live comparison identifies a concrete issue.
- [x] Capture depth 1, 5, and 8 in both Balanced and Performance at matching camera route, resolution, renderer, and fixture. Inspect the native-size frames for player/enemy separation, loot visibility, boss-warning contrast, warm torch pools, and the fog-of-war boundary.
- [x] Run the matched live performance candidate at the exact control topology and tiers from Task 1. Report frame interval p50/p95, process p95, draw calls, first-spawn cost, sample counts, hardware, renderer, viewport, and any vsync/cap noise. The pinned probe covers depth 1 only; no deep-floor performance inference is made. The final-source Performance gate missed by 0.128 ms; see as-built.
- [x] Run `make client-unit`, the relevant existing windowed bot scenarios, `make validate-shared`, and `make maintainability` after tuning. Do not put `make ci`/`make ci-full` in this worker plan; the coordinator owns the combined batch gate after integration.

### 5. Prepare the handoff

- [x] Write `docs/as-built/v505_dungeon-depth-mood.md` with catalog/schema changes, selected values, image links, exact commands, performance table, limits, and a separate statement for what screenshots versus measurements prove.
- [x] Report the exact base/dependency revisions, clean/dirty state, every changed/deleted/untracked path, all ignored evidence to preserve, focused test outcomes, and any files likely to conflict at integration.
- [x] Leave the worktree uncommitted. Do not run `/finish`, push, clean up, or edit the coordinator checkout.

## Verification command summary

```bash
make validate-shared
godot --headless --path client --script res://tests/test_dungeon_depth_mood.gd
godot --headless --path client --script res://tests/test_render_environment_presentation.gd
godot --headless --path client --script res://tests/test_dungeon_depth_lighting.gd
godot --headless --path client --script res://tests/test_dungeon_torch_placement.gd
make client-unit
make maintainability
make bot-visual scenario=dungeon_light_readability_shallow HEADLESS=0
make bot-visual scenario=dungeon_light_readability_boss_balanced HEADLESS=0
make bot-visual scenario=dungeon_light_readability_boss_performance HEADLESS=0
make bot-visual scenario=dungeon_light_readability_deep HEADLESS=0
ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_frame_pacing_probe HEADLESS=0
```

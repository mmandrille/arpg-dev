# v511 As-Built — Monster variant looks

- **Date:** 2026-10-01
- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)
- **Spec:** [v511 spec](../specs/v511_spec-monster-variant-looks.md) · **Plan:** [v511 plan](../plans/v511_2026-10-01-monster-variant-looks.md)
- **Baseline:** `425b9ae4`
- **Scope:** client presentation only. No server, protocol, rules, golden, replay, or asset-pack change.

## What shipped

- `shared/assets/kit_monster_presentation.v0.json` gained an additive top-level `variants` section (schema updated): `rarities` (tint, scale multiplier, eye glow, aura ring), `depth` (tint per biome palette id) and `families` (eye mesh name, aura radius scale) for the five dungeon families (three KayKit skeletons, Quaternius wolf and bat).
- `MonsterVariantResolver` (pure) resolves `(scene key, rarity, palette id)`; depth and rarity tints blend into one detail-layer overlay (strength `1-(1-d)(1-r)`). `MonsterVariantLook` applies it: shared cached materials with the detail layer (texture and `albedo_color` stay free for status/hit-flash owners), recolored model **eye glow** on the model's own `*_Eyes` mesh, a ground **aura ring**, and a scale multiplier on the model child (server `visual_scale` is untouched). Applied once at node creation from `main.gd::_make_entity_node`; the depth palette comes from `DungeonDepthLighting.palette_id_for_level`.
- Catalogued ordinary monsters get a neutral albedo base (the server `visual_tint` is ignored for them). Bosses, companions, `visual_model` overrides, training dummies and uncatalogued scenes are untouched (tested).
- `tools/validate_monster_variants.py` (wired into `validate_shared.py`) cross-checks rarity ids against `dungeon_generation.monster_rarities`, depth keys against `biome_palettes`, families against kit monster scenes, a neutral `common`, and distinct non-common looks.
- Death: the aura ring hides when a monster enters terminal death.

## Task 1 inventory findings

The three KayKit skeletons each have a separate `Skeleton_*_Eyes` mesh with its own `Glow` material; the bat has a separate `Eyes` mesh (palette material). The wolf has one mesh and no eye node. Eye glow therefore recolors existing meshes (no new geometry) for skeletons and bat; the wolf gets tint, scale and aura only (`eye_mesh: ""`).

## Deviations from the plan

- `ModelReactionController._capture_meshes` skips nodes carrying meta `presentation_skip_tint` (2 lines). Without it the controller's highlight/status passes disabled eye emission and overwrote the aura's albedo/alpha. `main.gd::_apply_model_tint` honors the same meta. Flag for v513 (same file).
- Palette is resolved once at node creation (rarity and depth do not change for a live monster); no re-apply path was needed.
- The `ModelTint` 64-entry cap needed no change: catalogued families use a constant white albedo base, and variant materials use their own 256-entry cache (`MonsterVariantLook.MAX_CACHED`). The full 5 x 4 x 3 sweep creates 76 materials.
- No regen-screenshots suite was added; a dedicated showme focus `monster-variants` (`--level`, `--grayscale`) was added instead to avoid touching the suite catalog shared by other slices.
- The real-camera scenario `v511_monster_variant_pack` captures the live benchmark arena (town level, so palette "" and common rarity: it proves the unchanged neutral path in the real camera). A live champion/rare in the real play camera was **not** obtained: generated-dungeon monsters are not near the spawn and bot navigation setup is disallowed by the movement contract. Champion/rare/unique and depth looks are proven through the same `_make_entity_node` path in the showme matrix.

## Evidence (ignored, local)

- Showme matrix (5 families x 4 rarities), depth 1/5/8 and a grayscale depth-5 variant: `.artifacts/v511-monster-variant-looks/matrix_L{1,5,8}.png`, `matrix_L5_gray.png`. Review: rarity ordering reads in grayscale through ring brightness/size (unique strongest, common none) and body brightness; champion vs rare in grayscale rely mainly on ring radius and the rare scale bump (weakest pair). Depth palettes shift the common skeleton hue (warm cave, violet halls, teal vault).
- Real-camera capture: `.artifacts/bot-captures/v511_monster_variant_pack_camera.png`.
- Matched live A/B (Apple M4 Pro, Godot 4.7.2 Forward+ Metal, 1920x1080, `dungeon_frame_pacing_probe`: level -1, 24 live monsters, seed `dungeon_levels_fast_247`; control = same code with `variants.families` emptied, interleaved pairs): `.artifacts/v511-perf/{balanced,performance}/`, raw logs and reports per run, aggregate via the plan's report tool.

| Metric | Control | Variants | Budget |
|---|---:|---:|---|
| Balanced first-spawn `_process` median (10 vs 9 valid) | 337.7 ms | 342.1 ms (+4.4) | <= max(20 ms, 10%) |
| Balanced first-spawn max | 422.4 ms | 430.8 ms (+8.4) | same |
| Balanced frame interval p95 median | 17.32 ms | 17.38 ms (+0.06) | <= max(0.5 ms, 5%) |
| Balanced process p95 median | 26.84 ms | 24.10 ms | same |
| Balanced draw calls / primitives p50 | 108 / 126,457 | 108 / 126,457 | <= max(10, 5%) |
| Performance (3 pairs) frame p95 / process p95 / draw calls | 17.18 / 17.88 / 68 | 17.26 / 17.42 / 68 | same |

One treatment run (balanced 4) was rejected by the report's own window-size check (`expected 1920, got 2324`), a host artifact; its bot run passed. Process-p95 maxima near 1,017 ms occur in both sides (one-off host spikes), so medians are reported.

## Limits

- The probe's steady camera does not frame the monsters, so draw calls/primitives identical to the control do **not** measure in-view cost. Expected in-view cost is bounded: shared materials, eye glow 0 extra draws, one extra draw per champion-or-above aura ring. First-spawn cost (all 24 nodes built, tinted, auras added) is measured and within budget. A weaker GPU remains untested.
- No live in-camera champion capture (see Deviations).
- `_apply_entity_visual_metadata` recomputes the base tint from the delta entity; a partial delta without `type` resolves to white, which matches the neutral base.
- Champion vs rare differ least in grayscale (ring radius/scale only); a tuning pass could widen aura alpha if review wants more separation.

## Focused verification

| Command | Result |
|---|---|
| `make validate-shared validate-assets` | PASS (2,243 + 451 checks) |
| `.venv/bin/pytest tools/test_validate_monster_variants.py tools/assets/test_validate_assets.py tools/test_validate_shared.py tools/test_scenario_movement_audit.py tools/test_validate_codemap.py -q` | 41 passed |
| `make client-unit` (includes `test_monster_variant_looks`, 8 checks; `test_kit_monsters`; `test_look_and_feel_polish`) | PASS |
| `make bot-client SCENARIO=v511_monster_variant_pack HEADLESS=1` | PASS (6.4 s) |
| `BOT_STEP_DELAY=0.0 make bot-visual scenario=v511_monster_variant_pack` | PASS, capture saved |
| `make maintainability` | PASS (`main.gd` 6,650 lines vs baseline 6,632 + 25) |
| `python3 tools/validate_codemap.py` | PASS |

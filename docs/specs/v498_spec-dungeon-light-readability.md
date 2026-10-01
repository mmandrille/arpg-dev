# v498 — Dungeon light and readability

- **Status:** Complete (combined `make ci` passed)
- **Date:** 2026-09-30
- **Codename:** `dungeon-light-readability`
- **Area:** Graphics
- **Depends on:** v497 frame-pacing baseline, so additional lighting cost is visible.
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D1, D7–D9.

## Purpose

Give shallow and deep dungeons a stronger, consistent atmosphere while keeping the player, enemies, loot, paths, and combat warnings easy to read. Tune the combined effect of biome palette, key light, torch pools, render fog, and the gameplay fog-of-war overlay from the actual player camera. The older deep-room capture is a useful warning, but live play is the acceptance surface.

## Non-goals

- New dungeon geometry, biome-generation rules, monster art, or a replacement fog-of-war system.
- New light/shader plugins or a cinematic post-processing pass that hides interactive information.
- A default-quality change or a GPU cost that erases the v497 gain.

## Acceptance criteria

- [ ] Before/after real-renderer captures use the live player camera and HUD in a shallow and deep dungeon, with enemies, loot, a telegraph, torch-lit and unlit space, and the fog-of-war overlay enabled.
- [ ] A visual review confirms that the player and active enemy silhouettes stay distinct from walls/floor, warning shapes are recognizable before impact, loot remains findable, and unexplored areas remain meaningfully obscured.
- [ ] Lighting has a visible near/far and shallow/deep hierarchy without clipping highlights or crushing interactable silhouettes. Catalog changes can be toggled and are bounded by schema validation.
- [ ] Balanced and Performance tiers both preserve interaction readability; feature differences between tiers are intentional and captured.
- [ ] Matched live-client runs show no material regression in frame p95, draw calls, or first-spawn cost against the v497 baseline. The plan defines the permitted budget before tuning.

## Scope and likely files

- **Presentation data:** `shared/assets/render_presentation.v0.json`, `shared/assets/fog_presentation.v0.json`, `shared/assets/dungeon_torch_presentation.v0.json`, and biome palette data, with schema updates only where needed.
- **Client:** `client/scripts/scene_lighting_rig.gd`, `client/scripts/dungeon_depth_lighting.gd`, `client/scripts/fog_of_war_overlay.gd`, `client/scripts/dungeon_torch_lights.gd` only if catalog-only tuning cannot achieve the goal.
- **Tests/docs:** loader/schema and quality-tier tests, real play-camera capture paths, as-built comparison.
- **Contracts:** no combat rule or protocol change expected.

## Test and bot proof

- `make validate-shared`, focused render/fog/torch tests, `make client-unit`, and existing line-of-sight/torch client scenarios.
- Use the real renderer for before/after captures; headless `gl_compatibility` checks do not prove Forward+ lighting, SSAO, glow, or fog appearance.
- Pair the visual gate with the same live performance topology used in v497.

## Asset/plugin decision

- **Borrow** existing kit torches, biome palettes, and in-repo render/fog catalogs.
- **Reject** new light assets, shader packs, and Godot plugins for this slice. Small in-repo shader changes require a separate measured reason in the plan.

## Open questions and risks

- A dark room capture can differ from live play because the fog-of-war overlay and player light are active there. Establish the live baseline before moving exposure or torch values.
- Brighter environments can flatten depth. Compare shallow and deep scenes side by side, and keep warnings readable before optimizing mood.

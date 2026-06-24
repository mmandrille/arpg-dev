# v331 As-Built — Hero Visibility Lighting

**Codename:** hero-visibility-lighting  
**Date:** 2026-06-23  
**Status:** Shipped

---

## What it proved

World-space dungeon fog works correctly in isometric mode (existing) and has been extended to
`chest_view` perspective mode via a physical `OmniLight3D` on the hero. The key discovery: a pure
canvas-overlay approach cannot achieve camera-angle-independent darkness in perspective mode without
depth-buffer access (which fails on macOS/Metal). The correct architecture for perspective is:

- **Dark ambient** (`directional_scale: 0.0, ambient_scale: 0.0` via `perspective_ambient_suppression`)
- **OmniLight3D** parented to `player_anchor` at 2 m height, `range = light_radius`, with shadow casting
- **Canvas fog transparent** in perspective mode (`visibility_perspective = 1.0`); isometric path unchanged
- **Shadow polygons hidden** in perspective mode (LOS polygon projections are wrong from low angles;
  OmniLight3D shadow casting replaces them for wall occlusion)

---

## Architecture decisions

### Canvas fog vs 3D light

| Approach | Tried | Outcome |
|----------|-------|---------|
| Ground-plane shader + height sampling | ✗ | Angle-dependent; near-horizontal rays always give FAR_WORLD_DIST |
| `hint_depth_texture` in canvas shader | ✗ | White screen on macOS/Metal — canvas depth not supported |
| Screen-space vignette (isometric mode for perspective) | ✗ | "Porthole" effect; user rejected |
| `OmniLight3D` + zero ambient | ✓ | Correct world-space light radius, camera-angle-independent |

### Character shadow

The `OmniLight3D` at 0.5 m (inside hero body) cast a large forward shadow. Fixed by raising to
`height_offset: 2.0` (above the character's head) and disabling shadow casting on `character_visual`
subtree via `set_perspective_camera()` → `GeometryInstance3D.SHADOW_CASTING_SETTING_OFF`.

### Ambient suppression routing

`main.gd` previously called `FogPresentationLoaderScript.ambient_suppression()` directly. Replaced
with `fog_overlay.ambient_suppression_params()` (1-line replacement, zero net lines). The overlay
returns `perspective_ambient_suppression` ({`directional_scale: 0.0, ambient_scale: 0.0`}) in
perspective mode and the original values for isometric, keeping both paths tunable in JSON.

### Emissive markers through walls

Unshaded/emissive `player_status_effect_markers.gd` 3D nodes bypass all lighting and appear through
OmniLight3D shadows. In perspective mode, `EliteAuraPreviewSync.sync()` now accepts
`(perspective, hero_pos, light_radius)` and culls aura markers for entities beyond `light_radius`.
Known limitation: entities within range but around a wall corner still show markers (OmniLight3D
shadows occlude the mesh but not emissive materials). Other marker types (burning, stun, etc.) are
not yet culled.

---

## Tuning surface (`shared/assets/fog_presentation.v0.json`)

All perspective-mode lighting is data-driven. No recompile needed for tuning:

```json
"point_light": {
  "energy": 3.0,          // OmniLight3D brightness
  "attenuation": 2.0,     // falloff steepness (higher = sharper edge)
  "color": "#ffffff",     // warm up with "#ffd6a0" for torchlight
  "range_multiplier": 1.0,// omni_range = light_radius * multiplier
  "height_offset": 2.0,   // metres above player_anchor (must exceed character height)
  "shadow_enabled": true  // false = cheaper but sees through walls
},
"perspective_ambient_suppression": {
  "directional_scale": 0.0,  // set >0 to add fill light beyond OmniLight3D range
  "ambient_scale": 0.0
}
```

---

## Files changed

| File | Change |
|------|--------|
| `client/scripts/fog_of_war_overlay.gd` | World-space shader, OmniLight3D management, `ambient_suppression_params()`, shadow polygon skip in perspective, character shadow toggle |
| `client/scripts/fog_presentation_loader.gd` | `point_light()`, `perspective_ambient_suppression()` accessors + defaults |
| `client/scripts/elite_aura_preview_sync.gd` | Perspective culling by distance via `_cull_far_aura_markers()` |
| `client/scripts/main.gd` | 3 line replacements (bind + 2× EliteAuraPreviewSync.sync calls), 0 net lines |
| `shared/assets/fog_presentation.v0.json` | `point_light` + `perspective_ambient_suppression` sections |
| `shared/assets/fog_presentation.v0.schema.json` | Schemas for both new sections |
| `client/tests/test_fog_of_war_overlay.gd` | Updated perspective falloff_mode assertion to `"point_light"` |
| `tools/bot/scenarios/client/83_camera_mode_setting.json` | Updated chest_view fog assertions |

---

## Open gaps

- Emissive status markers (burning, stun, pinning root, rogue mark) visible through walls in
  perspective mode — same structural issue as aura, not yet culled
- `OmniLight3D.shadow_enabled = false` disables wall occlusion; `shadow_enabled = true` is the
  default but is expensive — tune with `point_light.shadow_enabled` in JSON
- No LOS (line-of-sight) check for marker culling; distance-only approximation

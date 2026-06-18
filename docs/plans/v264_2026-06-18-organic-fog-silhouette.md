# v264 Plan - Organic Fog Silhouette

Status: Complete
Goal: Make the fog-of-war light/gloom/darkness boundary organic instead of perfectly circular.
Architecture: Extend the existing `FogOfWarOverlay` code-native shader with deterministic angular
edge noise that offsets the visual light and gloom thresholds without changing server-authoritative
radius values. Keep LOS shadow polygons as the opaque overlay layer above the organic mask, and
expose debug fields for unit and bot assertions.
Tech stack: Godot GDScript/CanvasItem shader, client unit tests, client bot scenario assertions.

## Baseline and Shortcut Decision

Builds on v253 fog radius, v255 LOS shadow mask, and v262 doorway LOS occlusion. This slice changes
only client presentation; no protocol/server visibility contract changes are needed.

Asset/plugin decision: reject external assets/plugins and imported fog art. Borrow the existing
`FogOfWarOverlay` shader and debug/bot assertion pattern.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/fog_of_war_overlay.gd` | Add deterministic organic edge shader parameters and debug state |
| Modify | `client/tests/test_fog_of_war_overlay.gd` | Prove organic-edge debug state and no-radius fallback |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Allow fog organic-edge bool/int/float assertions |
| Modify | `tools/bot/scenarios/client/67_fog_of_war_overlay.json` | Assert organic edge is enabled in base fog proof |
| Modify | `tools/bot/scenarios/client/68_fog_los_shadow_mask.json` | Assert LOS shadows coexist with organic edge |
| Modify | `docs/specs/v264_spec-organic-fog-silhouette.md` | Mark complete during close-out |
| Modify | `docs/progress/slice-lifecycle.md` | Add v264 lifecycle row |
| Add | `docs/as-built/v264_organic-fog-silhouette.md` | Record shipped behavior and proof |
| Modify | `PROGRESS.md` | Update current status and deferred scope |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines and grandfathered files stay under their
ratchet allowance.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd` is not touched.
- [x] `client/scripts/bot_assertion_handlers.gd` is grandfathered; keep the change narrow.
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Keep the organic mask in the existing focused fog overlay file.
- [x] Do not add a shader/plugin/asset pipeline.

Verification:
```bash
make maintainability
```

## Task 1 - Organic Fog Shader

Files:
- Modify: `client/scripts/fog_of_war_overlay.gd`
- Modify: `client/tests/test_fog_of_war_overlay.gd`

- [x] Step 1.1: Add deterministic angular noise helpers to the CanvasItem shader.
- [x] Step 1.2: Offset visual light/gloom thresholds while preserving debug radius values.
- [x] Step 1.3: Expose organic-edge debug fields and add unit assertions.

```bash
make client-unit
```

## Task 2 - Bot Proof

Files:
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Modify: `tools/bot/scenarios/client/67_fog_of_war_overlay.json`
- Modify: `tools/bot/scenarios/client/68_fog_los_shadow_mask.json`

- [x] Step 2.1: Add fog debug assertions for `organic_edge_enabled`, edge pixels, and segment count.
- [x] Step 2.2: Update base fog and LOS shadow scenarios to assert organic edge data.

```bash
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask
```

## Task 3 - Lifecycle Docs

Files:
- Modify: `docs/specs/v264_spec-organic-fog-silhouette.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v264_organic-fog-silhouette.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Mark spec and plan complete.
- [x] Step 3.2: Add as-built proof and lifecycle row.
- [x] Step 3.3: Update `PROGRESS.md` current status and deferred production fog-art scope.

```bash
make maintainability
```

## Final Verification

- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay`
- [x] `HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask`
- [x] `make maintainability`
- [ ] Autoloop final batch gate: `make ci`

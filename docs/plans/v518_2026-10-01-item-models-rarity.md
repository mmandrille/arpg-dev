# v518 — Item model thumbnails and stronger rarity cues

- **Status:** Complete — focused slice gates and combined batch `make ci` passed.
- **Spec:** [`v518_spec-item-models-rarity.md`](../specs/v518_spec-item-models-rarity.md)
- **Baseline:** `3a132626aa29581da6e3fdb36833c80e22b77b12`
- **Date:** 2026-10-01
- **Prerequisites:** Integrated v503/v508/v514/v515 presentation and rarity paths; no sibling slice dependency.

## Review gate

- Scope is client-only and retains the existing shared-model and manifest identity path; no wire, item mutation, or game-rule changes.
- Acceptance maps to focused tests, shared/asset validation, real-renderer captures, and the required inventory lab visual bot.
- There is no protocol/schema/golden work, server outcome, deterministic simulation, or gameplay tuning change. `equipment_display.v0.json` is presentation configuration for the existing model tint path.
- Asset decision is recorded in the spec: adopt the project-native renderer/cache, borrow existing manifest-backed models, reject new families/dependencies and per-slot 3D scenes.
- Likely shared-file overlap is limited to the rarity cue catalog, equipment display configuration, icon drawer, item-render helper, `loot_node_factory.gd`, and CODEMAP. Reconcile each path against the integrated tree before handoff.
- Security router classifies local manifest-backed presentation rendering as out of scope for its build workflow; there is no new untrusted resource path or input surface.

## Implementation sequence

### 1. Resolve eligible model thumbnails

- [x] Add a focused item-model thumbnail helper that resolves an item through existing item presentation data, hand-item mapping, and the manifest. Never construct `res://` paths from an item field alone; pass only the manifest's registered runtime resource.
- [x] Reuse a single offscreen render surface per active thumbnail service and cache final `Texture2D` thumbnails by resolved asset ID. Avoid creating a persistent SubViewport/3D scene for each slot. Render lazily only for models requested by the inventory/equipment view; unknown/unresolvable models fall back to existing 2D icon drawing.
- [x] Integrate the cached thumbnail into `ItemIconDrawer` for supported inventory/equipment item slots without changing slot button ownership, draw guards, blocked overlays, invalid-requirement overlays, labels, hotbar badges, or input hit targets. Other shared icon callers retain the 2D behavior by default.
- [x] Add a focused Godot test for representative hand and armor models, same-model cache reuse, manifest rejection/empty identity fallback, and existing icon fallback. Use fixtures derived from shared presentation/manifest catalogs rather than duplicating tuning/model data.

### 2. Strengthen and replace rarity indicators

- [x] Increase the shared equipment tint strength setting in `shared/assets/equipment_display.v0.json`; preserve equipped and ground code paths' existing rarity values and unknown-rarity behavior. Confirm the same setting governs rig-native equipment in both paths; if not, keep one data-owned setting as the source.
- [x] Replace slot rarity letter drawing with a distinct non-text geometric marker using each existing catalog shape. Replace world rarity letters with non-text 3D shapes using the catalog mapping, at the existing marker location/scale. Preserve the revealed full rarity name label and item display name.
- [x] Keep hue and all cue eligibility/suppression rules from `RarityCueLoader`. Preserve mystery offer concealment, category exclusions, empty/unknown rarity suppression, and item identity.
- [x] Extend tests for one-to-one shape distinction, no text-marker node, model tint strength use, and all existing suppression cases. Ensure missing rarity is never defaulted into a visible rarity cue.

### 3. Render and verify

- [x] Capture and inspect inventory/equipment thumbnails and grayscale rarity marks with the existing windowed showme infrastructure; preserve captures under `docs/as-built/assets/v518/`.
- [x] Capture ground rarity shapes/tint and run the `inventory_lab_drop_item` visible client bot; record bounded visual/render observations and inspect item identity and legibility.
- [x] Run the focused and repository client unit checks; run shared-data validation, asset validation, and `make maintainability`.
- [x] Run `make bot-visual scenario=inventory_lab_drop_item BOT_STEP_DELAY=0.05` after v520's windowed pass; scenario passed.
- [x] Write [`docs/as-built/v518_item-models-rarity.md`](../as-built/v518_item-models-rarity.md) with exact outcomes, capture links, bounded visual/performance limits, and any unmet criterion. Update the Assets and inventory rows in `docs/CODEMAP.md`; add v518 rows to lifecycle/codename indexes as the repo convention requires.
- [x] After the visual gate, review `git diff --check`, list all changed/untracked/ignored evidence, record dirty state, and report integration conflicts or none.

## Data-driven tuning and maintainability

The tint intensity belongs to the existing schema-backed `equipment_display.v0.json`; do not introduce new hardcoded color strength. The helper should stay small and independent of `inventory_panel.gd`; keep rendering and cache state out of the already-large panel. If a touched file is over its ratchet, do not grow it; extract only the minimum cohesive helper required by this slice. The thumbnail cache is a small bounded set keyed by registered asset ID, and UI slot redraws reuse its textures.

## Handoff boundary

This worker performs focused slice verification only. Do not run combined `make ci` or `make ci-full`, invoke `/finish`, commit, push, create a branch, or edit the coordinator's checkout. The coordinator compares the complete changed path list with integrated `main`, resolves overlaps, and runs the combined batch gate after all accepted slices are integrated.

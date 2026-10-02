# v518 — Item model thumbnails and stronger rarity cues

- **Status:** Complete — focused slice checks and combined batch CI passed.
- **Date:** 2026-10-01
- **Codename:** `item-models-rarity`
- **Accepted batch assignment:** v518
- **Baseline:** `3a132626aa29581da6e3fdb36833c80e22b77b12`
- **ADRs:** [ADR-0006](../adr/0006-asset-pipeline.md), [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md)
- **Dependencies:** v503's manifest-backed equipment models, v508's rarity cues, v514's UI theme, and v515's inventory slot/tooltip styling are integrated at the assigned baseline.

## Purpose

Inventory and equipment slots still use flat, code-drawn family icons even though matching KayKit and fallback models already appear on the floor and in the equipped character. Show those existing models as compact item thumbnails, and strengthen the existing rarity color treatment for equipped and ground models. Keep rarity legible without color by replacing the current rarity letters with distinct geometric shapes.

This is a client presentation slice. Item identity, interactions, rules, server outcomes, and asset provenance remain unchanged.

## Asset decision (ADR-0018)

- **Adopt:** the existing project-native GLB render and thumbnail path, backed by `assets/manifests/assets.v0.json`, `shared/assets/item_visuals.v0.json`, and item presentation definitions.
- **Borrow:** current KayKit hand models and the existing manifest-backed equipment fallback models for slot thumbnails and the already-supported ground/equipped presentations.
- **Reject:** new asset families, downloads, plugins, runtime asset fetching, generated duplicate model art, and a viewport/3D scene per slot.

## Scope

- Render a small inventory/equipment thumbnail from the existing model associated with an item's presentation. Reuse one renderer and cache the resulting texture by model asset ID; do not create a persistent viewport, scene, camera, or 3D model node per occupied slot. Keep a safe existing 2D icon fallback when an item has no supported registered model or render is unavailable.
- Increase the shared rarity tint strength used by equipped and ground equipment models. Keep tint presentation-only and derived from the item's existing rarity; do not infer a rarity that is absent.
- Replace text letters in slot and ground rarity cues with the already-established, distinct rarity shapes. Make the shape a non-text cue that remains distinguishable in grayscale. Preserve rarity colors as a redundant signal.
- Preserve cue allowlisting and suppression for empty/unknown rarity, non-equipment categories, and concealed mystery offers. Preserve item names and identity, tooltip text, prices, filters, drag/drop, equip, and pickup behavior.
- Record before/after real-renderer captures for slot thumbnails, grayscale rarity cues, and ground/equipped tint visibility. Keep capture fixtures bounded to representative model families and rarities.

## Non-goals

- Changes to item definitions, gameplay rarity/drop rates, network payloads, server authority, persistence, colliders, model geometry, materials, asset manifests, or inventory interactions.
- Replacing current fallback icon drawing for unsupported or unregistered items.
- Full colorblind-accessible redesign or new rarity tiers; geometric rarity cues remain supplemental to the existing rarity name in text contexts.
- Rendering models independently for every bag/shop/stash/market/blacksmith row. This slice owns inventory and equipment slots; shared icon callers keep their existing behavior unless the common inventory/equipment draw path covers them without expanding renderer cost.
- Performance claims beyond confirming the bounded renderer/cache design and noting any measured costs. No busy-scene benchmark is implied.

## Acceptance criteria

1. Supported item families display thumbnails rendered from the same manifest-backed models used by floor/equipped presentation. Unsupported items retain their current family icon, and the thumbnail retains the item and slot's existing interactions.
2. Thumbnail rendering is bounded: at most one reusable offscreen 3D rendering surface exists per active inventory presentation owner; completed thumbnails are cached by resolved model identity, so occupied slots do not each instantiate a renderer or retain a live 3D model.
3. Equipped and ground equipment models use a stronger rarity tint than the baseline, using the shared presentation setting. Item identity and all game outcomes stay unchanged.
4. Inventory/equipment and ground rarity cues contain no rarity letters and use distinct non-text shapes for every supported rarity. Each shape remains distinguishable without hue in grayscale captures.
5. Empty or unknown rarity and unsupported categories continue to suppress rarity cues. Concealed mystery offers remain unidentified; their silhouette behavior does not change.
6. Focused unit checks cover model-to-thumbnail resolution, cache reuse/fallback, cue mapping/suppression, and the configured tint strength. The real-renderer evidence shows representative supported slot models, grayscale slot cues, stronger equipped/floor tint, and no loss of item identity.
7. `make bot-visual scenario=inventory_lab_drop_item` passes, preserving the inventory/drop/pickup flow. Data and asset validators pass. No server, protocol, or gameplay-rule surface changes.

## Security assessment

The security skill classifies this local, client-only presentation change as out of scope for security implementation guidance: it adds no network/API, persistence, authorization, secret, untrusted HTML, or user-controlled resource-path surface. Model lookup remains restricted to existing shared presentation data and the local asset manifest.

## Likely surfaces

| Area | Files |
|---|---|
| Thumbnail lookup/render/cache | `client/scripts/item_icon_drawer.gd`, a small new model-thumbnail helper, `client/scripts/item_rules_loader.gd` or `client/scripts/item_visuals_loader.gd` |
| Inventory/equipment | `client/scripts/inventory_panel.gd`, `client/scripts/inventory_panel_styles.gd` only if needed |
| Rarity | `client/scripts/rarity_cue_presenter.gd`, `shared/assets/rarity_cues.v0.json`, rarity catalog tests |
| Model tint | `shared/assets/equipment_display.v0.json`, `client/scripts/equipment_display_loader.gd`, equipped/ground model tint paths and tests |
| Visual/bot proof | `client/scripts/showme/`, `tools/showme/`, `tools/bot/scenarios/client/` only if focused capture requires a fixture |
| Documentation | `docs/as-built/v518_item-models-rarity.md`, `docs/CODEMAP.md`, `docs/progress/slice-lifecycle.md`, `docs/progress/slice-codename-index.md` |

## Focused verification

- Relevant focused Godot unit scripts for item thumbnail rendering, inventory, rarity cues, loot factory, and equipped model tint.
- `make validate-shared`, `make validate-assets`, and `make maintainability`.
- `make regen-screenshots SUITE="item-icons"` (or a narrower supported focus), inspect representative before/after and grayscale captures.
- `make bot-visual scenario=inventory_lab_drop_item BOT_STEP_DELAY=0.05`.
- `git diff --check`.

## Integration risks

`item_icon_drawer.gd` is shared by several commerce and storage panels, and `rarity_cue_presenter.gd` serves both 2D slots and world loot; keep changes targeted and compare each affected call path. Inventory UI, ground models, and rarity cues may intersect the sibling UI slices only through integrated baseline assets; no sibling feature changes are expected. The coordinator runs combined `make ci` after all accepted slices are integrated.

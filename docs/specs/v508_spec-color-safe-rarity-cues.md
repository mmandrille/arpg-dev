# v508 — Color-Safe Rarity Cues

- **Status:** Complete; coordinator integration and combined batch `make ci` passed. The final no-aura refinement needs a fresh matched cost sample.
- **Date:** 2026-10-01
- **Batch baseline:** `5365832b9029e0d9e178d57f2a95a6a697b01acc` (v500 completed)
- **Dependencies:** v503 ground gear and v504 quest/badge ground loot must be integrated before this slice edits shared ground-item presentation paths.
- **Architecture:** ADR-0001 D2 (server-owned items and loot), ADR-0018 D9 (real-renderer visual gate).

## Purpose

Item rarity currently relies heavily on hue: ground model tint, glow and beam, ground-label color, and the backgrounds/borders of inventory, stash, shop, and market item slots. The inventory, stash, shop, and market detail text already includes `Rarity: ...`, but that text is not always visible while scanning. Make the five equipment rarities—common, magic, rare, unique, and set—recognizable from a compact shape and/or text cue alongside the existing color, both on the ground and in item browsing surfaces.

## Scope and non-goals

- Add one schema-backed client presentation vocabulary for each rarity: a full readable name, a distinct compact symbol/shape, and any visual sizing needed by the cue. Keep existing rarity colors unless real-camera contrast review shows an adjustment is needed.
- Show a persistent, non-color cue next to dropped equipment at the play camera without forcing hidden loot labels to stay visible. When the ground label is revealed, include the full rarity name without repeating a rarity prefix already present in an item display name. Preserve hover, reveal, and loot-filter behavior.
- Show dropped items directly on the floor without the surrounding aura, spawn ring, pickup beam, or backdrop pad. This visual simplification applies to equipment, gold, consumables, quests, and badges; the small equipment rarity letter remains separate from the item shape.
- Show the compact cue on occupied equipment slots in inventory/paper doll, stash, shop/vendor, market, and blacksmith staging where rarity color is currently used. A full rarity name remains available in an item detail or tooltip. Existing colors, icons, affordability, requirement warnings, and item interactions retain their meanings.
- Do not apply equipment rarity cues to gold, consumables, quest items, badges, empty slots, or unrevealed mystery offers. These have category or unknown-state presentation of their own. The cue must not expose hidden rarity or other unidentified attributes.
- Do not change rarity balance, loot rules, item identity, prices, server state, protocol, persistence, replay, gameplay, filter thresholds, model sourcing, or broad UI layout.

## Observable acceptance criteria

1. All five item rarities have distinct text and silhouette cues in a schema-backed presentation catalog. The catalog's keys match the rarities in `shared/rules/item_templates.v0.json`; duplicate or missing compact cues fail validation. Unknown/empty rarity never silently displays as common.
2. At normal and maximum playable camera zoom, visible dropped equipment has a persistent cue readable without hue, including on dark/light ground. Alt reveal adds the full rarity word to its label. Five rarities are distinguishable in grayscale captures, and a crowded drop still respects the existing label cull/filter/hover rules. No dropped item draws a floor aura, ring, beam, or background pad.
3. Occupied equipment slots in inventory, stash, vendor, market, and blacksmith staging show the compact cue at the actual slot sizes. The icon, item count/label, blocked/invalid marker, and selected/hover state remain legible; the full rarity name is available in the existing or augmented detail text.
4. Category items and mystery offers show no misleading rarity cue. After identification/reveal, a real rarity can be shown through the same cue path. No server-provided item fields are rewritten for presentation.
5. Focused Godot tests derive expected cues from the catalog and exercise all five rarities, missing/unknown values, item-category exclusions, and mystery-offer exclusion. Existing inventory/loot interaction scenarios pass.
6. Before/after real-renderer captures from the playable camera and UI captures cover all five rarities, normal/max zoom, dark/light ground, grayscale inspection, small slots, and warning/hover variants. Evidence records readable and unreadable cases, contrast measurements or sampled luminance ratios for cue foreground/background, and the renderer/quality tier. Check no material frame-pacing or draw-call regression on a matched loot fixture if world markers add draw work.

## Likely surfaces

| Area | Existing paths to inspect or adapt |
|------|-----------------------------------|
| Presentation data | `shared/assets/` new rarity cue JSON and schema; `shared/rules/item_templates.v0.json` only as a rarity-key reference, never a balance edit |
| Ground loot | `client/scripts/loot_node_factory.gd`, `loot_label_filter.gd`, `loot_label_hover.gd`, and v503/v504's integrated ground presentation paths |
| Item UI | `client/scripts/item_icon_drawer.gd`, `inventory_panel.gd`, `stash_panel.gd`, `shop_panel.gd`, `market_panel.gd`, `blacksmith_panel.gd`, `item_tooltip_panel.gd`; preferably a small reusable cue presenter rather than duplicate mappings |
| Validation/tests | `tools/validate_shared.py` or focused presentation validator; `client/tests/test_loot_node_factory.gd`, `test_inventory_panel.gd`, relevant shop/stash/market tests, one focused cue test |
| Proof/docs | `skills/showme` existing `floor-item`, `inventory`, `shop`, `stash`, and market focuses or a focused extension; `docs/as-built/v508_color-safe-rarity-cues.md` and captures |

## Verification and visual proof

- `make validate-shared`; focused Godot cue, ground-loot, inventory, shop/stash/market tests; `make client-unit` and `make maintainability` after implementation.
- Client bot: `make bot-client SCENARIO=inventory_lab_drop_item HEADLESS=1` and the affected stash/vendor/market scenarios for flow preservation. For human visual verification, run `make bot-visual scenario=inventory_lab_drop_item` (the exact scenario ID is confirmed against the integrated scenario catalog before execution).
- Real-renderer captures through `/showme`, the existing `floor-item` suite, and a focused rarity comparison capture for the playable camera and item UI (`inventory`, `shop`, `stash`, market, and blacksmith where a focus exists or is added). Current `regen-screenshots` suites do not include those panel focuses, so do not list `inventory` as a suite without registering it. Keep matched before/after and grayscale inspection. A headless bot pass is not readability proof.
- The worker performs focused checks only. The coordinator runs combined `make ci` after integrating all accepted slices.

## Dependencies, integration risks, and decisions

- v503 owns ground equipment presentation and v504 owns quest/badge drops; their integrated data and code are the implementation baseline for ground cues. Spec and plan may proceed now, but code editing in overlapping ground files waits until those slices integrate. Recheck paths, tests, and captures at that point.
- `loot_node_factory.gd` and UI coordinators have existing consumers and file-size ratchets. Keep cue logic in a focused module; extract instead of growing over-limit panels. Do not make a shape that obscures an item's pickup target or alter collision.
- **Adopt:** existing in-repo Godot drawing, Label3D, item icons, and screenshot harness. **Borrow:** the current rarity palette and schema-backed shared-assets pattern. **Reject:** new third-party art, fonts, colorblind plugins, external asset pipeline, and runtime downloads. No external research is needed for this bounded cue design.
- **Security router:** this is a fixed-catalog visual presentation change. It does not add input parsing, auth, network or file-input behavior; the documentation-only stage is outside the security workflow.

## Planning gate

The `/plan` review must map every acceptance criterion to a focused check or real-renderer capture, confirm exclusion behavior and marker placement, inspect v503/v504 integration drift, and resolve any material overlap before implementation.

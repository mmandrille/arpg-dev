# v508 Plan — Color-Safe Rarity Cues

- **Status:** Complete; integrated and combined batch `make ci` passed. The user's final no-aura refinement still needs a matched cost sample.
- **Date:** 2026-10-01
- **Spec:** [`v508_spec-color-safe-rarity-cues.md`](../specs/v508_spec-color-safe-rarity-cues.md)
- **Recorded batch base:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **Prerequisites:** v503 ground gear and v504 quest/badge ground loot integrated before editing overlapping ground presentation files. Recheck the integrated HEAD, changed paths, scenario catalog, and screenshots before implementation.

## Review gate and ownership

The spec passes the plan review: the outcome is a client-only presentation change; all five authoritative rarity values are already carried in item/loot state; no protocol, schema of network payloads, Go logic, golden fixture, world preset, loot weight, price, or replay change is needed. Acceptance criteria map to checks in the final table. The server continues to own item rarity and pickup outcomes. A new presentation catalog owns cue labels, shapes, and sizes; it must not be interpreted as loot rules. The existing category label/filter and mystery-offer concealment remain authoritative for what the client may display.

Code execution is intentionally held at this gate. v503/v504 may change `loot_node_factory.gd`, item presentation data, tests, or screenshot fixtures. Read those integrated files and update this plan's exact edit points if they drift; do not merge an old ground implementation over their work. No material product decision remains open in the accepted brief.

**Asset/plugin decision:** adopt in-repo Godot `Label3D`, CanvasItem drawing, existing icons and screenshot harness; borrow current rarity colors and `shared/assets` schema pattern; reject outside assets, fonts, plugins, and runtime downloads. A small code-native neutral foreground/outline cue suits the KayKit view and needs no new manifest entry or external research note. Geometry used to construct each catalog-selected shape is rendering code; cue dimensions, symbol and label mapping live in data.

**Security router:** fixed catalog and visual drawing only. The planned implementation introduces no new network, authentication, file-input, or user-controlled HTML/rendering surface. Security workflow remains out of scope unless implementation scope changes.

## File map and shared-file risk

| Area | Planned path and ownership |
|------|----------------------------|
| Cue data | New `shared/assets/rarity_cues.v0.json` and `.schema.json`: exact five rarity entries, full/compact labels, distinct shape IDs, sizing and neutral outline/foreground parameters. `shared/rules/item_templates.v0.json` is read only for key parity. |
| Catalog loading and drawing | New focused `client/scripts/rarity_cue_loader.gd` and `rarity_cue_presenter.gd` (static guarded load and normalized shape/badge drawing); `client/scripts/item_icon_drawer.gd` receives an optional item rarity for occupied equipment icons. No autoload. |
| Ground | `client/scripts/loot_node_factory.gd`: non-pickable marker bound to an equipment loot node and full rarity word on revealed label; preserve v503/v504 model/category handling. Touch `loot_label_filter.gd` or `loot_label_hover.gd` only if the longer label changes their measured hit/cull behavior. |
| Inventory and trade | Minimal call-site wiring in `client/scripts/inventory_panel.gd`, `stash_panel.gd`, `shop_panel.gd`, `market_panel.gd`, `blacksmith_panel.gd`, and `blacksmith_item_craft_slot.gd` as required. Existing `item_tooltip_panel.gd` and panel detail text supply the full label; add it where missing (notably blacksmith). Keep mystery offers on `_draw_mystery_icon` with no rarity argument. |
| Tests and validation | Focused `client/tests/test_rarity_cues.gd`, existing `test_loot_node_factory.gd`, `test_inventory_panel.gd`, `test_shop_panel.gd`, `test_stash_panel.gd`, market/blacksmith tests, a focused presentation-data validator and test registered through `tools/validate_shared.py`, `scripts/client_smoke.sh`. |
| Capture and docs | Existing `/showme` focuses in `skills/showme/`; if required, add a `rarity-cues` fixture/focus and suite in `client/scripts/showme/`, `skills/showme/scripts/render_focus.py`, and `tools/showme/screenshot_catalog.py`. Update `docs/CODEMAP.md`; write `docs/as-built/v508_color-safe-rarity-cues.md` with captures under `docs/as-built/assets/v508/`. Coordinator alone updates `PROGRESS.md` and lifecycle after integration. |

`loot_node_factory.gd`, ground item data and tests are probable v503/v504 overlaps. Make no ground edit until the coordinator confirms their integrated result and a current comparison. UI files may be independent, but this assigned session stops before any implementation regardless.

## Maintainability and presentation data

At the recorded base, `inventory_panel.gd` is 1623 lines against a 1623-line baseline; `stash_panel.gd` 1194/1194; `shop_panel.gd` 1247/1243; `market_panel.gd` 1081/1059; `blacksmith_panel.gd` 793/771. Recount after dependencies integrate. Keep new logic in the focused cue files and make touched grandfathered panels end at or below their recorded baseline by extracting a coherent item-icon/rarity presentation helper, especially in panels already above baseline. Do not collapse lines cosmetically or pass full panel namespaces into an extracted module. `loot_node_factory.gd` (464 lines) and `item_icon_drawer.gd` (243) are below 600 at this base; do not turn either into a new over-limit file. Run `make maintainability` after edits.

The catalog holds the cue mapping and display sizing, with a schema and validator enforcing exact key parity and distinct shape/compact-label combinations. Keep the existing color palette unless a measured readability fix is necessary. No gameplay tuning value is introduced. Code geometry may use normalized coordinates within a catalog-sized marker because it is rendering mechanics, not balance tuning. Do not copy current loot weights or rarity probabilities into tests.

## Ordered implementation tasks (after prerequisite integration)

### 1. Rebase the evidence and capture the before state

Paths: integrated `client/scripts/loot_node_factory.gd`, item presentation catalogs, `client/scripts/showme/`, `tools/showme/screenshot_catalog.py`, existing unit and client bot scenarios.

- [x] Compare v503/v504 changed paths and behavior with this file map; resolve any overlap before editing. Recheck loot categories, item name prefixing, mystery offers, slot sizes, and file-size baselines.
- [x] Capture a five-rarity **play-camera** fixture at normal and maximum zoom over dark/light terrain, plus inventory, stash, shop, market and blacksmith representative slots at actual UI size. Use a deterministic fixture with known rarity fields; record renderer, quality tier, zoom, viewport, item IDs, and lighting. The capture-only cue-off state is a matched visual baseline, not a separate pre-v508 executable.
- [x] Record a matching world-loot frame-time/draw-call baseline if a persistent marker will add render objects. Preserve the same fixture, host, renderer, tier, warmup and sample count for after comparison.

Smallest check: inspect resulting PNGs and metadata; `make regen-screenshots-list` for registered suite names. No headless claim of visual readability.

### 2. Define and validate one rarity cue vocabulary

Paths: new `shared/assets/rarity_cues.v0.json`, `.schema.json`, validator and validator test, new `client/scripts/rarity_cue_loader.gd`, `client/tests/test_rarity_cues.gd`.

- [x] Define each rarity's full word, short badge text, distinct silhouette ID, foreground/outline treatment, and presentation sizes for world and slot contexts. Keep the badge readable on grayscale/dark/light backgrounds and distinguish common as positively as the higher tiers. Expose no cue for empty/unknown rarity.
- [x] Validate exact key parity with `item_templates.v0.json` rarities; reject duplicate short text or indistinguishable silhouette selections. Load from one guarded static client reader, with no parallel hardcoded rarity-to-shape maps.
- [x] Test five mapped rarities, unknown/missing rarity, case normalization, schema failures, and runtime loader behavior using catalog-derived expectations.

Verify: `make validate-shared`; `.venv/bin/pytest -q tools/test_validate_rarity_cues.py` if a focused validator is added; `godot --headless --path client --script res://tests/test_rarity_cues.gd`.

### 3. Present a non-color cue on ground equipment

Paths: integrated `client/scripts/loot_node_factory.gd`, `loot_label_hover.gd` and `loot_label_filter.gd` only if necessary, `client/tests/test_loot_node_factory.gd` and focused label tests.

- [x] Attach a small non-pickable, neutral outlined letter marker to equipment loot, positioned clear of the model and pickup collider. Use the catalog letter and size. No marker on gold, consumables, badges or quest objects.
- [x] Remove the floor aura around every drop at the user's request: no rarity torus, spawn ring, pickup beam, or fallback background pad. Keep the item model or item-shaped fallback visible on the floor and retain revealed labels and pickup behavior.
- [x] When the label is revealed, include the full rarity word from the catalog. Detect an existing leading rarity word to avoid a duplicate; keep the underlying `display_name` and entity unchanged. Preserve filter thresholds, Alt reveal, hover hit area, crowded label cap, and text clipping.
- [x] Test all five marker identities and label text, category exclusions, unknown rarity, collision/pickup invariance, label cull/filter behavior, and no direct mutation of the input dictionary.

Verify: `godot --headless --path client --script res://tests/test_loot_node_factory.gd`; `godot --headless --path client --script res://tests/test_loot_label_filter.gd`; `make bot-client SCENARIO=inventory_lab_drop_item HEADLESS=1`. The historical `loot_label_filter` scenario named in v376 as-built is absent from this base, so extend the focused client test or add a dedicated extended-tier scenario if the integrated code requires interaction-level filter proof. Reinspect new markers through the play camera.

### 4. Add the compact badge to equipment browsing

Paths: `client/scripts/item_icon_drawer.gd`, affected panel call sites, `client/scripts/rarity_cue_presenter.gd`, focused Godot tests.

- [x] Draw the compact text/shape badge in a reserved icon corner so the item-family glyph, shop price, count, affordability dim, warning state, and hover/focus state remain legible. Reuse the same presenter across occupied equipment slots in inventory/paper doll, stash, vendor, market listing/staging, and blacksmith staging. Do not infer rarity from icon family or color; pass the item's actual rarity. No badge for mystery, non-equipment, or empty slots.
- [x] Ensure each surface's detail/tooltip gives the full rarity name. Retain existing item names, tooltip stats, and selection behavior. For the blacksmith, add a full text rarity line if its current detail omits it.
- [x] Keep panel ratchets by extracting a coherent existing icon/row block where needed. Test each surface at minimum slot size, hover, blocked/invalid, unaffordable, empty, category and mystery cases.

Verify: focused Godot tests for `test_rarity_cues.gd`, `test_inventory_panel.gd`, `test_shop_panel.gd`, `test_stash_panel.gd`, `test_market_listing_rows.gd`, `test_blacksmith_panel.gd`; `make bot-client SCENARIO=town_vendor_shop_panel HEADLESS=1`; `make bot-client SCENARIO=account_stash_panel HEADLESS=1`; `make bot-client SCENARIO=market_board_ui HEADLESS=1`.

### 5. Prove readability and close the slice handoff

Paths: `client/scripts/showme/`, screenshot harness only if needed, `docs/as-built/assets/v508/`, `docs/as-built/v508_color-safe-rarity-cues.md`, `docs/CODEMAP.md`.

- [x] Capture matching after views at normal/max play-camera zoom, dark/light ground, all five rarities, and the affected UI contexts. Inspect normal-color and grayscale versions side by side at native screen size. Sample cue foreground/background luminance on the actual captured backgrounds and record ratios and any failure/adjustment; explicitly check small badges, crowded drops, and warning/hover states.
- [x] If the cue adds persistent world draw objects, compare matched before/after frame p50/p95, draw calls, sample counts and fixture details. Reduce marker complexity or instances and repeat if it introduces a meaningful regression. Do not infer performance from screenshots.
- [x] Run relevant screenshot suite after changes and inspect PNGs under `.artifacts/screenshots/latest/`. Commit representative before/after images only through the coordinator's eventual closeout; preserve ignored capture evidence for handoff.
- [x] Run final focused tests, `make validate-shared`, `make client-unit`, and `make maintainability`; list every changed/untracked path and ignored evidence, commands and outcomes for the coordinator. Write the as-built with the remaining limits. Do not run `make ci`, `make ci-full`, commit, push, or modify coordinator `main` in this batch slice session.

Human visual command: `make bot-visual scenario=v508_rarity_cue_preset_camera` to inspect a server-provided magic rarity through the real client. The existing `inventory_lab_drop_item` fixture omits rarity and verifies the no-cue boundary. The dedicated scenario remains extended-tier.

## Acceptance-to-proof map

| Spec criterion | Focused proof |
|----------------|---------------|
| 1. Five distinct catalog cues | Schema + cross-catalog validator negative cases; `test_rarity_cues.gd` |
| 2. Ground cue and revealed full label | `test_loot_node_factory.gd`, `test_loot_label_filter.gd`, client `inventory_lab_drop_item` scenario and a focused filter scenario only if needed; real-camera dark/light, normal/max, crowded, grayscale captures |
| 3. Compact UI cue and full detail | Focused inventory/shop/stash/market/blacksmith tests and client flow scenarios; actual-size UI captures including hover/warning |
| 4. Category/mystery exclusions | Cue helper and ground/panel negative tests; mystery vendor capture; no authoritative payload mutation assertion |
| 5. Deterministic focused tests | `make validate-shared`, focused Godot gates, `make client-unit`, flow scenarios |
| 6. Readability and render cost | Matched before/after images, grayscale and luminance log, p50/p95 and draw calls if markers add work, with fixture/renderer/sample count |

## Integration handoff

Leave implementation and focused verification in this detached worktree after prerequisites are integrated into its baseline. Report the complete changed/untracked file list and ignored screenshot/measurement evidence to the coordinator. The coordinator compares paths and intended behavior against the integrated result, runs combined `make ci` after all accepted slices, writes lifecycle closeout, and commits. This plan contains no worker `make ci` gate.

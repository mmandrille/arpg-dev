# v508 — Color-safe rarity cues

**Status:** Integrated; focused checks and combined batch `make ci` passed (11m41s, 2026-10-01). The user's final no-aura refinement postdates the matched cost sample, so final-state performance remains unverified.

## Implementation

- A schema-backed `shared/assets/rarity_cues.v0.json` defines full names, compact letters, five distinct slot shapes, one-letter world markers, and presentation sizes/colors. The cross-catalog validator requires exact parity with gameplay rarity keys and rejects duplicate names, letters, shapes, and world symbols.
- `RarityCueLoader` allowlists equipment from shared item definitions and suppresses cues for unknown/empty rarity, categories outside equipment, and concealed mystery offers. `RarityCuePresenter` draws a neutral outlined badge in occupied item slots and a persistent `Label3D` marker on dropped equipment. Revealed ground labels add the full rarity word only when the display name does not already lead with it.
- Inventory, stash, shop, market, and blacksmith pass their item dictionaries to the shared icon drawer. The blacksmith detail now includes the full rarity name. A blocked inventory slot retains its rarity in the tooltip. No item payload or gameplay rule is changed.
- On user review, ground drops were simplified to the item model or item-shaped fallback alone. The factory no longer adds rarity torus glow, spawn ring, pickup beam, or fallback background pad. The small equipment rarity letter and revealable label remain; gold, consumables, quests, and badges keep their distinct item shapes without an aura.
- A focused five-drop fixture uses the playable isometric camera configuration at normal/max zoom, two controlled ground tones, the existing client lighting, 60 warmup frames, and 120-frame samples. The capture-only baseline flag hides the new cues for matched comparisons. A new visual bot captures a known magic preset through the actual client camera. Focused UI fixtures show blacksmith staging and a concealed mystery offer.

## Focused verification

| Check | Result | Scope |
|-------|--------|-------|
| `make validate-shared` | PASS | 2,215 checks plus CODEMAP |
| `make validate-assets` | PASS | 451 asset checks |
| `.venv/bin/pytest -q tools/test_validate_rarity_cues.py tools/test_showme.py` | PASS | 7 tests |
| `make client-unit` | PASS | All client unit gates after final edits; rarity cue test: 54 assertions |
| `make maintainability`, `git diff --check` | PASS | File-size, extraction, and whitespace gates |
| Client bots: `inventory_lab_drop_item`, `town_vendor_shop_panel`, `account_stash_panel`, `market_board_ui` | PASS | One scenario each, no failures |
| `make bot-visual scenario=v508_rarity_cue_preset_camera BOT_STEP_DELAY=0.05` | PASS | Magic drop and inventory cue in the actual 1920×1080 play camera |
| `make bot-visual scenario=inventory_lab_drop_item BOT_STEP_DELAY=0.05` | PASS | Missing-rarity starter item correctly shows no cue |
| `make regen-screenshots SUITE=floor-item` | PASS | Final 60/60 captures, representative 640×480 item inspected |
| `make regen-screenshots SUITE=item-icons` | PASS | 1/1 capture |
| Post-review `godot --headless --path client --script res://tests/test_loot_node_factory.gd` | PASS | 798 assertions, including no aura on equipment, gold, quests, and badges |
| Post-review `godot --headless --path client --script res://tests/test_item_visuals.gd` | PASS | Existing item model resolution checks; exit-time resource warnings remain |
| Post-review `make regen-screenshots SUITE=floor-item` | PASS | 60/60 captures with the aura removed |

Direct focused Godot tests also passed for rarity cues, loot node factory, loot label filter, inventory, shop, stash, market listing rows, and blacksmith. The direct shop/stash scripts and mystery-shop capture printed Godot exit-time RID/ObjectDB cleanup warnings despite passing sentinels and exit code 0; `make client-unit` passed.

## Visual evidence

Representative screenshots are in [assets/v508](assets/v508/); all four normal/max and dark/light pairs, JSON samples, and more UI captures remain under `.artifacts/showme/v508/`. At 1920×1080, the five-drop fixture shows C, M, R, U, S without hue at both zooms. See [normal dark before](assets/v508/ground-normal-dark-baseline.png), [normal dark after](assets/v508/ground-normal-dark-after.png), [max light after](assets/v508/ground-max-light-after.png), and [normal light grayscale](assets/v508/ground-normal-light-after-gray.png). The [revealed rare label](assets/v508/ground-max-light-reveal.png) gives the full word above the persistent marker row. An exploratory forced-all-label capture overlapped neighboring drops; the final capture reveals one center label, while existing label filter and hover tests cover culling.

The preceding five-drop images were captured before the user's aura-removal refinement. Current [five-rarity floor](assets/v508/rarity-cues-no-aura.png), [640×480 sword](assets/v508/floor-item-long-sword-no-aura.png), [gold](assets/v508/gold-no-aura.png), and [quest leaf](assets/v508/quest-leaf-no-aura.png) show the item shapes and separate equipment rarity letters with no surrounding geometry. The 798-assertion loot factory test checks the absence of all four former aura node types across modeled gear, primitive equipment, gold, quests, and badges.

Native-size [inventory](assets/v508/inventory-after.png), [stash](assets/v508/corpse-inventory-after.png), [vendor](assets/v508/shop-after.png), [market](assets/v508/market-offer-after.png), and [blacksmith](assets/v508/blacksmith-after.png) captures show occupied equipment badges. The inventory includes small slots, a blocked item, and a full-word tooltip. The [mystery vendor](assets/v508/mystery-shop-after.png) retains its unidentified silhouette without a rarity badge. [Actual camera ground](assets/v508/play-camera-magic-ground.png) and [actual camera inventory](assets/v508/play-camera-magic-inventory.png) confirm the positive path; the [missing-rarity starter sword](assets/v508/play-camera-missing-rarity.png) confirms suppression. At 640×480, the [floor item suite sample](assets/v508/floor-item-long-sword-640.png) has a small but legible M, less prominent than in the native-size fixture.

For a bounded contrast sample, the normal-zoom screenshot ground pixels were RGB `(46,46,38)` on dark terrain and `(192,178,155)` on light terrain. A cue-region grayscale sample reached 216/255; the darkest outline sample was 0/255 on dark and about 14/255 on light. WCAG relative-luminance contrast from these sampled values is **9.60:1** for the light letter on dark ground and **9.27:1** for the dark outline on light ground. The opposite pairings are weak (1.53:1 and 1.46:1), which motivates the outline. These fixture samples do not cover every map material or display.

## Matched render sample

Five long swords, one per rarity, were rendered at 1920×1080 on Apple M4 Pro / Metal, Godot 4.7.2 Forward+, balanced quality, with identical lighting/camera/ground. Each baseline and after capture had 60 warmup frames and 120 samples. Baseline is a capture-only cue-off view of integrated v503/v504 code, not a separate pre-v508 executable. Frame values are milliseconds; draw calls are median.

These matched samples predate the aura-removal refinement; their draw-call counts should not be read as measurements of the final no-aura renderer.

| Zoom / ground | Baseline p50 / p95 | After p50 / p95 | Draw calls baseline → after |
|---------------|--------------------|-----------------|-----------------------------|
| Normal / dark | 16.651 / 17.705 | 16.657 / 17.638 | 25 → 35 |
| Normal / light | 16.709 / 17.555 | 16.642 / 17.647 | 25 → 35 |
| Max / dark | 16.688 / 18.045 | 16.684 / 17.428 | 25 → 35 |
| Max / light | 16.632 / 17.806 | 16.692 / 17.649 | 25 → 35 |

Two-glyph ground markers first added 20 draw calls; a single outlined letter reduced that to 10 for five drops. The short, vsync-bound fixture shows no material frame-pacing shift. It does not establish busy-scene performance or end-to-end gameplay FPS. Bot screenshot startup FPS overlays are not performance samples.

## Integration notes and limits

- v503 ground-equipment and v504 quest/badge files in this worktree are inherited prerequisites. `loot_node_factory.gd`, `docs/CODEMAP.md`, `skills/showme`, and bot scenarios overlap; the coordinator must compare each with the integrated state.
- Real-camera positive proof uses one magic preset. Five-rarity, zoom, terrain, grayscale, and matched cost proof use the deterministic play-camera fixture. Blacksmith staging and mystery vendor proof use focused UI fixtures.
- When the server omits rarity, the client intentionally shows no cue. No inferred common rarity or mystery identity is exposed. This was observed on a world starter sword and its inventory copy.
- Combined `make ci` passed on the integrated batch. `make ci-full` was not run. No batch worker commit or push was made.

Human visual command: `make bot-visual scenario=v508_rarity_cue_preset_camera`.

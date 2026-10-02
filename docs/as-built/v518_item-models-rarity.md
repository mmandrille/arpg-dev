# v518 — Item model thumbnails and stronger rarity cues (as built)

- **Status:** Focused implementation, data, windowed visual, and bot gates passed; combined batch CI remains coordinator-owned.
- **Date:** 2026-10-01
- **Spec / plan:** [spec](../specs/v518_spec-item-models-rarity.md) · [plan](../plans/v518_2026-10-01-item-models-rarity.md)
- **Base:** `3a132626aa29581da6e3fdb36833c80e22b77b12`
- **Scope:** Client presentation only. Item identity, inventory operations, authoritative outcomes, model assets, and item rules are unchanged.

## What changed

- `ItemModelThumbnailCache` resolves equippable items through the existing item presentation catalog, hand-visual catalog, and asset manifest. It renders each registered asset into a 96×96 image through one lazily created SubViewport, caches by asset ID (maximum 32), and releases each live model after capture. No renderer or live model is created per slot. Unavailable models retain the existing 2D family icon.
- Only `InventoryPanel` opts into model thumbnails. Other `ItemIconDrawer` callers preserve their existing 2D behavior unless they explicitly request a model.
- Shared `equipment_display.v0.json` raises rig-native rarity tint strength from 0.25 to 0.45. The existing equipped and ground model paths consume this same setting; rarity colors and missing-rarity behavior are unchanged.
- Rarity catalog entries retain their full names and unique shapes. Slot cues draw filled geometric shapes; ground cues use one-surface two-tone mesh silhouettes. Letter/world-symbol fields were removed from the catalog and schema. Revealed full rarity names remain available in labels/tooltips, and `RarityCueLoader` filtering/suppression is preserved.

## Focused verification

| Check | Result |
|---|---|
| `make client-unit` | PASS — full headless client unit suite, including thumbnail and rarity gates |
| `godot --headless --path client --script res://tests/test_item_model_thumbnail_cache.gd` | PASS — representative manifest-backed models, cache-key reuse, fallback exclusions, cache cap, and tint config |
| `godot --headless --path client --script res://tests/test_rarity_cues.gd` | PASS — latest 70 assertions, including configured marker scale and prior suppression/exclusion cases |
| `godot --headless --path client --script res://tests/test_loot_node_factory.gd` | PASS — 813 assertions, including the configured rig-native rarity tint |
| `godot --headless --path client --script res://tests/test_inventory_panel.gd` | PASS — 17 assertions |
| `godot --headless --path client --script res://tests/test_item_icon_drawer.gd` | PASS |
| `python tools/validate_shared.py` using the existing project virtualenv | PASS — 2,257 checks |
| `python tools/validate_codemap.py` using the existing project virtualenv | PASS |
| `python tools/assets/validate_assets.py` using the existing project virtualenv | PASS — 464 checks |
| `pytest -q tools/test_validate_rarity_cues.py` | PASS — 3 tests |
| `make maintainability` | PASS — file-size, extraction-coupling, and progress-dashboard ratchets |
| `make bot-visual scenario=inventory_lab_drop_item BOT_STEP_DELAY=0.05` | PASS — visible client run, 1 passed / 0 failed |
| Combined batch `make ci` | PASS — 11/11 stages, 7m50s; `make ci-full` was not run |
| `bash -n scripts/client_smoke.sh` / `git diff --check` | PASS |

The bot runner created the worktree-local ignored `.venv` for its dependencies. No dependency was installed outside that environment. The earlier validation commands used the existing `/Users/mmandrille/git/arpg-dev/.venv` interpreter while reading this worktree's files.

## Windowed visual evidence

- [Inventory with model thumbnails and rarity marks](assets/v518/inventory.png) — equipped slots show manifest-backed 3D thumbnails; non-equipment bag items retain their 2D family icons.
- [Grayscale inventory](assets/v518/inventory-grayscale.png) — slot marks remain visible without color.
- [Ground rarity shapes on dark ground](assets/v518/rarity-cues-final-dark.png) and [baseline without cues](assets/v518/rarity-cues-baseline-dark.png) — all five shapes are visible at normal camera zoom and the full item labels remain suppressed until revealed.
- Sidecar capture samples record 17 draw calls with five world markers and 12 in baseline mode. This fixture shows one added draw call per marker in this scene. It is a bounded renderer observation, not a frame-rate or general performance claim.
- Marker size is data-owned at 0.045 after visual inspection found 0.09 too large for the normal gameplay camera.

Inventory capture command: `python3 skills/showme/scripts/render_focus.py --focus inventory`. Ground capture command: `python3 skills/showme/scripts/render_focus.py --focus rarity-cues --ground-tone dark`. The required `inventory_lab_drop_item` visible bot passed, covering the live inventory/drop interaction path. Focused captures are deterministic fixtures; they do not claim broad gameplay or performance proof.

## Handoff state and limits

The detached worktree is uncommitted and remains based on the assigned SHA. Tracked `.glb.import` metadata regenerated by Godot is restored. Ignored runtime output includes `client/.godot/`, generated client `.uid` files, `.pytest_cache/`, `.artifacts/showme/`, and the bot-created `.venv/`. No source integration conflict is known. The combined `make ci` remains the coordinator's batch gate.

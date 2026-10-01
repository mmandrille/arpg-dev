# v494 Plan — Dungeon room dressing

Status: Complete; combined `make ci` passed in 10m33s.
Goal: Add repeatable, restrained KayKit props to generated dungeon rooms without implying blocked routes.
Architecture: A pure planner selects roomy floor cells from the authoritative wall layout, filters them against walls, passages, interaction anchors and a conservative spawn-to-interactable route reserve, then applies catalog-owned density, weights, yaw and scale. The renderer builds one MultiMesh per selected asset at the wall-layout boundary; level teardown removes the entire dressing root. The session seed and level form the floor key. No server gameplay state changes.
Tech stack: shared JSON/schema, Godot GDScript, Python asset validation and capture tooling.

## Baseline and shortcut decision

v493 is complete. Reuse `DungeonKitFloor.plan_cells`, `KitPieceLibrary` and the existing wall renderer lifecycle. **Borrow** the vendored CC0 KayKit barrel, stacked barrels, crates and decorated table already in the manifest. **Reject** external packs and plugins. The spec's nonblocking default is adopted. Snapshot `entities` supply anchors immediately; level-change entity updates supply them before deferred wall construction. If an anchor is unavailable, wall and passage clearances remain conservative and the planner omits placements in uncertain areas.

## File map

| Action | Path | Responsibility |
|---|---|---|
| Modify | `shared/assets/dungeon_kit_presentation.v0.json` and schema | Data-owned placement limits, props and clearances |
| Modify | `tools/assets/validate_assets.py` | Resolve every dressing asset ID to a vendored environment entry |
| Modify | `client/scripts/dungeon_kit_presentation_loader.gd` | Return dressing catalog |
| Add | `client/scripts/dungeon_room_dressing.gd` | Pure placement planner and batched visual builder |
| Modify | `client/scripts/wall_renderer.gd` | Build/clear dressing with the wall layout |
| Modify | `client/scripts/main.gd` | Supply seed, level and static anchors at snapshot/transition boundary |
| Add | `client/tests/test_dungeon_room_dressing.gd` | Determinism, limits, exclusions and teardown |
| Modify | `shared/rules/worlds.v0.json` | Deep generated-floor lab for live renderer proof |
| Add | `tools/bot/scenarios/client/107_dungeon_room_dressing_deep.json` | Bounded deep-floor client scenario |
| Modify | `scripts/client_smoke.sh`, `scripts/bot_client.sh`, `client/scripts/surface_material_room_capture.gd` | Test registration, performance logging and real-renderer scene capture |
| Modify | `docs/CODEMAP.md`, `docs/as-built/v494_dungeon-room-dressing.md` | Ownership and evidence |

## Maintenance ratchet

Target: new source/test/tool files stay at or below 600 lines. `wall_renderer.gd` is currently 549 lines; keep additions small. `main.gd` is grandfathered and receives only short boundary calls. `validate_assets.py` is grandfathered and receives only an asset ID list extension. Defer extracting those coordinators because the slice changes only their existing call sites.

Verification: `make maintainability`.

## Task 1 — Catalog and baseline

- [x] Capture pre-change sparse and dense dungeon images and draw-call measurements with the real renderer; set an explicit instance and draw-call budget.
- [x] Add schema-backed `dressing` fields for depth bands, grid sampling, clearances, instance cap, prop weights/yaw/scale and their manifest IDs.
- [x] Extend asset validation so unknown or wrong-type IDs fail.

Verify: `make validate-shared && make validate-assets`.

## Task 2 — Planner and lifecycle

- [x] Implement deterministic cell ranking and weighted prop selection from seed/level, walls, static anchors and catalog. Reserve wall margins, narrow openings, the spawn-to-interactable corridor and spacing between props; expose a no-safe-candidate result.
- [x] Build a bounded set of MultiMeshes from cached kit meshes, with no per-frame reconstruction or collision.
- [x] Integrate build/clear with wall layout and level changes; pass snapshot/transition anchors at the appropriate boundary.
- [x] Add focused tests for identical/different keys, weights/toggles, clearances, cap, empty floor and teardown.

Verify: `make client-unit` (plus direct focused GDScript test while iterating).

## Task 3 — Gameplay and visual proof

- [x] Run the existing `wall_floor_dungeon_rollout` client scenario and dungeon navigation/combat scenarios without changing their CI tiers.
- [x] Run `make regen-screenshots SUITE=scenes` and inspect sparse/deep dungeon PNGs.
- [x] Capture before/after real-renderer play-camera frames on sparse and dense generated floors, compare draw calls and review route and combat-cue legibility. The pre-change live shallow floor measured 86 draw calls; keep the increase within **8 draw calls** and **32 prop instances** per floor, with matched captures to document the actual delta. Dedicated loot overlap was not captured and is recorded as a visual limit in the as-built.

Visual replay for the owner: `make bot-visual scenario=wall_floor_dungeon_rollout`.

## Task 4 — Documentation and final verification

- [x] Update CODEMAP and write the as-built with exact capture paths, before/after metrics, test results and remaining limits.
- [x] Leave PROGRESS and lifecycle closeout for the coordinating task in the owner's main checkout; the isolated worktree did not commit or transfer.

Final verification in this isolated worktree: `make maintainability`, `make validate-shared`, `make validate-assets`, `make client-unit`, and focused bot scenarios. The coordinating task runs one combined `make ci` after integration. No new branch, commit or `/finish` in this worktree.

## Deferred scope

Server-authored blockers, route changes, new props, asset downloads, interactive behavior and v495 hitch work remain separate.

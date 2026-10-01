# v504 — Quest and badge ground loot

**Status:** Integrated; focused checks and combined batch `make ci` passed (11m41s, 2026-10-01). Gameplay keys remain deferred until a real key item/drop path exists.

## What changed

- Extended schema-backed ground presentation data for `quest_leaf`, four distinct quest trophies, `upgrade_shard`, and `renew_stone`. The four wallet badges retain their item palettes and scales and now use the revised family medallion shape. Icon presentation and item identity are unchanged.
- Added `LootQuestBadgeShapes`, a bounded, client-authored low-poly shape helper selected only by eight explicit catalog shape names. The existing `LootNodeFactory` GLB resolver, primitive fallback, rarity effects, labels, and pickup root remain on their previous paths.
- Appended catalog-derived coverage to the existing factory test without removing v503's transferred gear coverage. Added pre-pickup `capture_frame` steps to the existing quest and blacksmith client scenarios. No new gameplay scenario or CI-pack entry was introduced.
- Updated `docs/CODEMAP.md` for the new helper. No manifest, GLB, external pack, plugin, or asset-pipeline change was needed.

## Verification

| Check | Result | Evidence / limit |
|---|---|---|
| `make validate-shared` | PASS | 2,208 shared checks; catalog/schema and CODEMAP valid. |
| `godot --headless --path client --script res://tests/test_loot_node_factory.gd` | PASS | 661 passed, 0 failed after an initial `godot --headless --path client --import` populated the fresh worktree's class registry. Repeated after visual refinements. |
| `make client-unit` | PASS | Full client unit suite passed after initial implementation; the focused factory test was rerun after the final mesh/palette refinements. |
| `make maintainability` | PASS | File-size, extraction-coupling, and dashboard gates; final rerun in handoff. |
| `make bot-client SCENARIO=quest_town_turn_in HEADLESS=1` | PASS | 1 scenario, 0 failures; quest pickup and turn-in retained. |
| `make bot-client SCENARIO=blacksmith_renew_item HEADLESS=1` | PASS | 1 scenario, 0 failures; renew-stone pickup retained. |
| `make bot scenario=steward_hunt_quest` | PASS | Existing protocol trophy-pickup scenario. |
| `make bot-visual scenario=quest_town_turn_in` | PASS | Reviewed player-camera leaf capture, preserved below. |
| `BOT_STEP_DELAY=0.0 make bot-visual scenario=blacksmith_renew_item` | PASS | Reviewed player-camera renew-stone capture, preserved below. The plain default-delay command timed out twice at the existing `bishop_debug_loot_dropped` wait, before the added capture step; its filtered log remains under `.artifacts/v504/debug/blacksmith_renew_item-client.log`. |
| `make regen-screenshots SUITE="floor-item"` | PASS | 60/60 existing floor-item equipment regression captures in `.artifacts/screenshots/20261001-121327/`. This suite does not enumerate quest or badge items. |
| `git diff --check` | PASS | Final handoff check. |

The first focused Godot command on a fresh checkout reported unresolved global classes in its log despite exit status 0. Importing the project resolved that local initialization issue; the repeated test reported 661/0. This first attempt is not counted as a pass.

## Reviewed render evidence

Player-camera proof: [quest leaf](assets/v504/quest-leaf-gameplay.png) and [renew stone](assets/v504/renew-stone-gameplay.png). The leaf is visible on its rarity tile near the hero but remains small in the dark quest lab; the stone is clear beside the player. Neither has observed floor clipping or label obstruction.

Focused real-renderer captures: [leaf](assets/v504/quest-leaf-focused.png), [wolf heart](assets/v504/wolf-heart-focused.png), [bat wing](assets/v504/bat-wing-focused.png), [archer head](assets/v504/archer-head-focused.png), [mob skull](assets/v504/mob-skull-focused.png), [respec badge](assets/v504/respec-badge-focused.png), [upgrade shard](assets/v504/upgrade-shard-focused.png), and [renew stone](assets/v504/renew-stone-focused.png). The head and skull were rotated and their face details improved after first-pass captures showed plain rounded silhouettes; the leaf palette was brightened after its first live-camera review. Final focused images were reviewed after those refinements.

No pre-slice quest/badge PNG was found. The first-pass and final renders document local iteration, not a full pre-slice baseline comparison. Focused renders prove the selected meshes are visible in the renderer; screenshots do not prove gameplay quality, performance, or every item's visibility in a live combat world. The player-camera captures cover one quest item and one badge-family resource, while trophy pickup is covered by the protocol scenario.

## Security and integration notes

The security skill router classified this trusted-catalog presentation work as outside its application-security workflow. Selection is constrained by the shared `groundShape` enum and an explicit client shape whitelist. No user-controlled text or paths, network endpoint, authentication, persistence, dependency, or server authority changed. The client helper creates render meshes only.

The coordinator transferred v503's 23-path uncommitted overlay into this worktree. v504 appended assertions to the shared `client/tests/test_loot_node_factory.gd` file and left v503's four tracked changes and 19 untracked files intact. The coordinator must compare the integrated result with both slices before cleaning worktrees. Generated `.glb.import` defaults from local Godot import were restored to the base revision and are not part of the handoff.

The shared Godot/bot runner is released. The combined `make ci` gate passed on the integrated batch. Keys remain deferred until a real gameplay item/drop path exists.

## Handoff file manifest

Base revision: `5365832b9029e0d9e178d57f2a95a6a697b01acc`. There is no v503 transfer commit; its uncommitted 23-path overlay was copied into this detached worktree. No path was deleted. The final worktree has 10 modified tracked paths and 33 untracked paths, with `client/tests/test_loot_node_factory.gd` shared between the two slices.

v503 transferred tracked modifications (4):

```text
client/tests/test_item_visuals.gd
client/tests/test_loot_node_factory.gd
shared/assets/equipment_display.v0.json
tools/bot/scenarios/client/10_full_equipment.json
```

v503 transferred untracked files (19):

```text
docs/as-built/assets/v503/amulet-after.png
docs/as-built/assets/v503/amulet-before.png
docs/as-built/assets/v503/belt-after.png
docs/as-built/assets/v503/belt-before.png
docs/as-built/assets/v503/boots-after.png
docs/as-built/assets/v503/boots-before.png
docs/as-built/assets/v503/chest-after.png
docs/as-built/assets/v503/chest-before.png
docs/as-built/assets/v503/gloves-after.png
docs/as-built/assets/v503/gloves-before.png
docs/as-built/assets/v503/ground-gear-camera-after.png
docs/as-built/assets/v503/ground-gear-camera-before.png
docs/as-built/assets/v503/helm-after.png
docs/as-built/assets/v503/helm-before.png
docs/as-built/assets/v503/ring-after.png
docs/as-built/assets/v503/ring-before.png
docs/as-built/v503_ground-gear-models.md
docs/plans/v503_2026-10-01-ground-gear-models.md
docs/specs/v503_spec-ground-gear-models.md
```

v504 tracked modifications (7, including the shared test):

```text
client/scripts/loot_node_factory.gd
client/tests/test_loot_node_factory.gd
docs/CODEMAP.md
shared/assets/item_presentations.v0.json
shared/assets/item_presentations.v0.schema.json
tools/bot/scenarios/client/75_quest_town_turn_in.json
tools/bot/scenarios/client/blacksmith_renew_item.json
```

v504 untracked files (14):

```text
client/scripts/loot_quest_badge_shapes.gd
docs/as-built/assets/v504/archer-head-focused.png
docs/as-built/assets/v504/bat-wing-focused.png
docs/as-built/assets/v504/mob-skull-focused.png
docs/as-built/assets/v504/quest-leaf-focused.png
docs/as-built/assets/v504/quest-leaf-gameplay.png
docs/as-built/assets/v504/renew-stone-focused.png
docs/as-built/assets/v504/renew-stone-gameplay.png
docs/as-built/assets/v504/respec-badge-focused.png
docs/as-built/assets/v504/upgrade-shard-focused.png
docs/as-built/assets/v504/wolf-heart-focused.png
docs/as-built/v504_quest-badge-ground-loot.md
docs/plans/v504_2026-10-01-quest-badge-ground-loot.md
docs/specs/v504_spec-quest-badge-ground-loot.md
```

Ignored local evidence and runtime state: `.artifacts/bot-captures/v504_quest_leaf_camera.png`, `.artifacts/bot-captures/v504_renew_stone_camera.png`, the first/final focused renders under `.artifacts/showme/`, the 60-image regression set and index under `.artifacts/screenshots/20261001-121327/`, and `.artifacts/v504/debug/blacksmith_renew_item-client.log`. The reviewed final PNGs are also preserved in `docs/as-built/assets/v504/`; keep that directory during integration. `.venv/` and `client/.godot/` are ignored local test/runtime caches, not source handoff. Generated `.glb.import` changes were restored to `HEAD`.

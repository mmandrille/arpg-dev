# v504 — Quest and badge ground loot

- **Status:** Complete; coordinator integration and combined batch `make ci` passed.
- **Date:** 2026-10-01
- **Codename:** `quest-badge-ground-loot`
- **Baseline:** `5365832b9029e0d9e178d57f2a95a6a697b01acc` (v500 batch base)
- **Dependency:** v503's 23-path uncommitted file overlay was transferred into this worktree. Its gear-pose and test changes are present, and the coordinator approved the v504 plan before implementation.
- **Scope decision:** Use existing quest/trophy and badge-family drops only. Defer keys until a gameplay key item and drop path exist.

## Purpose

Quest and badge-family drops currently use simple code-built primitives. Quest items include `quest_leaf` and four `quest_trophy_*` items; the `badge` presentation family is also used by `upgrade_shard`, `renew_stone`, and the four wallet badges. These pickups are easy to overlook and the quest and badge families are visually close to generic resource drops.

Give the existing quest and badge-family ground drops clear, readable silhouettes that fit the KayKit-led world presentation. Keep presentation catalog-driven and preserve the existing server-owned drop, pickup, wallet, and resource-bag behavior.

## Scope

- Improve floor presentation for every item definition that resolves to the `quest` or `badge` family in `shared/assets/item_presentations.v0.json`.
- Keep the existing item definitions, drop sources, rewards, pickup interactions, collision reach, and inventory/resource-wallet semantics unchanged.
- Use the existing `ItemRulesLoader` → `LootNodeFactory` presentation path, with v503's ground-gear changes preserved. Add item-specific ground-shape overrides where family-level sharing would make trophies or resources indistinguishable.
- **Keys are deferred.** The current item catalog and asset manifest have no key identity or key model; this slice does not add a visual-only prop or gameplay drop path.

## Non-goals

- Adding or changing server gameplay, item IDs, quest progress, reward rules, drop rates, auto-pickup, or persistence.
- Changing rarity colors, labels, pickup beams, loot filtering, or floor collision/pickup range.
- Importing a new asset pack, adding a plugin, or changing the asset pipeline without a revised adopt/borrow/reject decision in the approved plan.
- Implementing a key-consuming door/quest mechanic. If the user requests a gameplay key, split or expand that scope explicitly before planning.
- Adding a key item identity, key model, or synthetic key prop before the gameplay path exists.

## Asset decision

- **Adopt:** the existing item-presentation registry and `LootNodeFactory` primitive composition path; preserve the `EquipmentDisplayLoader`/manifest-backed model path used by ground gear and tuned by v503.
- **Borrow:** the KayKit-led low-poly scale and material language already present in the checked-in environment/loot assets.
- **Reject:** new external packs, plugins, and asset-pipeline dependencies for this slice. The checked-in shield-badge, stone chunks, forest plants, and whole-monster assets do not represent these distinct pickups. The plan selects small client-authored, low-poly mesh compositions through the existing loot-root path and records the exact item-to-shape mapping.

## Acceptance criteria — existing drops

- [ ] Every catalog item resolved to the `quest` or `badge` family selects a ground shape, and focused renders show distinct families. Normal gameplay-camera readability remains only partially proven: the live-camera leaf is small in the dark quest lab, and not every item has a live-camera capture.
- [x] The `quest_leaf` and `quest_trophy_*` items remain identifiable in focused captures; item-specific ground overrides give each trophy a distinct shape.
- [x] The `upgrade_shard`, `renew_stone`, and wallet badges resolve through their existing item identities; pickup scenarios passed.
- [x] No new asset reference was added. Ground shapes, scales, and tints are selected in the schema-backed catalog; fixed mesh topology is client presentation code.
- [x] The existing rarity glow, label, pickup beam, loot root, and fixed pickup collider remain intact in the unchanged factory path and focused tests.
- [x] Tests derive covered IDs from the shared catalogs and prove every eligible item builds its selected shape; unrelated primitive fallback and transferred v503 gear-model tests pass.
- [x] Existing quest, resource, and trophy pickup scenarios pass. Existing client scenarios now capture quest and resource drops before pickup.
- [x] Reviewed real-renderer captures show quest and badge-family drops in focused and player-camera views, with no clipping or label obstruction observed. Gameplay-camera proof covers the leaf and renew stone, not each trophy or wallet badge.

## Likely surfaces

| Area | Candidate files |
|---|---|
| Shared presentation | `shared/assets/item_presentations.v0.json`, `shared/assets/item_presentations.v0.schema.json`; possible item-specific entries under `items` |
| Client ground rendering | `client/scripts/loot_node_factory.gd`; focused `client/scripts/loot_quest_badge_shapes.gd` helper; preserve v503's gear model resolver/poses |
| Tests | `client/tests/test_loot_node_factory.gd`; retain v503's `client/tests/test_item_visuals.gd` coverage; focused shared-data validation only if schema/catalog constraints need extending |
| Bot/visual proof | Add pre-pickup captures to `tools/bot/scenarios/client/75_quest_town_turn_in.json` and `tools/bot/scenarios/client/blacksmith_renew_item.json`; retain `steward_hunt_quest` protocol pickup proof; `skills/showme/scripts/render_focus.py` supports `--focus floor-item --items <ids>` |
| Documentation | `docs/CODEMAP.md` only if new scripts or responsibilities are added; as-built and lifecycle entries are coordinator closeout work |

## Verification targets

- Focused Godot test: `godot --headless --path client --script res://tests/test_loot_node_factory.gd`.
- Client test suite: `make client-unit`.
- Shared presentation validation: `make validate-shared`.
- Asset validation when manifest or GLB references change: `make validate-assets`.
- Real-renderer visual capture: `python3 skills/showme/scripts/render_focus.py --focus floor-item --items quest_leaf,quest_trophy_wolf_heart,quest_trophy_bat_wing,respec_badge,stat_badge` (run after the coordinator reserves the shared Godot runner).
- Bot proof: `make bot-visual scenario=quest_town_turn_in` and `make bot-visual scenario=blacksmith_renew_item` for representative player-camera drops; `make bot scenario=steward_hunt_quest` for trophy pickup. The plan gives headless counterparts. All Godot/bot runs await the coordinator's shared-runner reservation.
- The coordinator runs combined `make ci` after all batch slices are integrated; no per-slice full CI.

## Security assessment

This slice changes only client-side presentation selection for trusted catalog item IDs and does not add user-controlled text rendering, input handling, persistence, networking, or authorization. The `meli-security-expert` router classifies this product/presentation-only scope as out of scope for a security workflow; no security-specific code change is indicated.

## Deferred follow-up

- A future gameplay-key slice may define a key item identity and drop path, then add its ground presentation through the same catalog-driven resolver.

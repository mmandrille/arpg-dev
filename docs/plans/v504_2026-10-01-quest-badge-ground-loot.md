# v504 Plan — Quest and badge ground loot

**Status:** Complete; integrated and combined batch `make ci` passed in 11m41s.

Spec: [`docs/specs/v504_spec-quest-badge-ground-loot.md`](../specs/v504_spec-quest-badge-ground-loot.md)

## Goal and baseline

Give current quest/trophy and badge-family floor drops clear, readable silhouettes, while keeping all authoritative item/drop/pickup behavior unchanged.

- Assigned baseline and current detached `HEAD`: `5365832b9029e0d9e178d57f2a95a6a697b01acc`. The coordinator transferred v503's exact 23-path uncommitted overlay (four modified tracked paths and 19 untracked paths); there is no new commit SHA to record.
- Current catalog scope: `quest_leaf`, four `quest_trophy_*` items; the `badge` family covers `upgrade_shard`, `renew_stone`, and the four `*_badge` wallet resources.
- Keys are deferred until a gameplay key item and drop path exist; do not add a synthetic prop or item identity in v504.
- Prerequisite: v503's ground-gear pose, tests, scenario, and evidence overlay is present and has been inspected. Its as-built says the production resolver was already available from v487/v490; v503 added seven gear poses, not a new resolver or quest model. Do not edit the transferred v503-owned paths until this plan passes coordinator review.

## Scope and asset decision

Presentation only. Keep server rules, item identities, reward generation, pickup authority, resource-bag/wallet behavior, rarity, labels, filters, and colliders unchanged.

- Adopt the existing catalog-driven `ItemRulesLoader` → `LootNodeFactory` path. The transferred v503 work confirms the resolver already selects manifest-backed gear models through `3d_model` and `EquipmentDisplayLoader.ground_pose_for`; v504 leaves that path untouched and uses the existing primitive-mesh fallback branch for quest and badge items.
- Borrow the existing KayKit low-poly scale/material language.
- Reject new external packs, plugins, and asset-pipeline dependencies. The in-repo `kaykit_weapon_adv_shield_badge_v0` is an equipment shield, not a wallet badge; Resource Bits stone chunks are environment rocks, not distinct quest trophies; forest bush/grass and whole monster GLBs do not provide detached leaves, wings, hearts, or heads. Reusing those assets would give several drops the wrong identity. Use small client-authored, low-poly mesh compositions in the existing loot-root fallback path; no `3d_model`, manifest, GLB, or `equipment_display` change is planned. Compare captures beside KayKit loot to review style coherence under ADR-0018.

### Exact ground-shape mapping

`ItemRulesLoader` copies a complete item-level `ground` override over its family `ground` object; it does not merge individual shape or tint keys. Keep shape, color, accent, and scale together in schema-backed `item_presentations.v0.json`. The shape names below are renderer-owned presets; catalog data owns item selection, palette, and overall scale. Preset mesh topology is fixed presentation construction, with no gameplay tuning or collision effect.

| Item ID(s) | Family | Ground preset | Data change |
|---|---|---|---|
| `quest_leaf` | `quest` | `leaf` | Keep family mapping; improve the current flat leaf composition through the new focused shape helper. |
| `quest_trophy_wolf_heart` | `quest` | `heart` | Add item-level ground override with a warm red palette. |
| `quest_trophy_bat_wing` | `quest` | `wing` | Add item-level ground override with a dark violet palette. |
| `quest_trophy_archer_head` | `quest` | `hooded_head` | Add item-level ground override; hood silhouette separates it from the bare skull. |
| `quest_trophy_mob_skull` | `quest` | `skull` | Add item-level ground override with a bone palette. |
| `respec_badge`, `stat_badge`, `skill_badge`, `resurrection_badge` | `badge` | `badge` | Keep existing item-level colors/scales; improve the medallion composition at family level. |
| `upgrade_shard` | `badge` | `shard` | Keep existing item-level colors/scales; replace only its ground shape with a faceted shard. |
| `renew_stone` | `badge` | `stone` | Keep existing item-level colors/scales; replace only its ground shape with a rounded stone. |

The current `LootNodeFactory` can already compose primitive meshes and preserve the rarity root, label, glow, beam, and fallback behavior. Its current `leaf` and `badge` constructions are too plain, and it has no trophy, shard, or stone presets. Extend that mechanism through a small independent quest/badge shape helper instead of routing these items to unrelated GLBs. Only `groundShape` schema enum values for the six new presets are needed; icon shapes remain unchanged.

## File map and overlap

| Candidate path | Responsibility | Overlap / ownership |
|---|---|---|
| `shared/assets/item_presentations.v0.json` + `.schema.json` | Item-level ground overrides above and six new allowed ground presets; no icon, gameplay, or GLB mapping change | No v503 edit to these files; retain its gear presentation resolution |
| `client/scripts/loot_node_factory.gd` | Dispatch the accepted quest/badge presets to the focused helper, preserving current loot-root effects and the existing GLB resolver | v503 did not edit this file; keep it under its current 464-line baseline and below 600 lines |
| `client/scripts/loot_quest_badge_shapes.gd` (new) | Independently build small, low-poly leaf, heart, wing, hooded-head, skull, medallion, shard, and stone meshes | v504-owned new presentation helper; no import of `LootNodeFactory` |
| `client/tests/test_loot_node_factory.gd` | Catalog-derived coverage for all 11 quest/badge items and unchanged fallbacks | v503 added gear coverage here; append assertions without removing or rewriting it |
| `client/tests/test_item_visuals.gd` | No planned edit | Preserve transferred v503 integration assertions; use the focused factory test instead |
| `tools/bot/scenarios/client/75_quest_town_turn_in.json` | Capture the existing ground `quest_leaf` before its pickup, then retain the authoritative turn-in proof | v504 adds one `capture_frame` step; no movement or new lab world |
| `tools/bot/scenarios/client/blacksmith_renew_item.json` | Capture the bishop-forced ground `renew_stone` before its pickup, then retain the resource-bag proof | v504 adds one `capture_frame` step; no server/debug behavior change |
| `docs/CODEMAP.md` | Add the new quest/badge shape helper to loot-presentation ownership | Shared docs may overlap with other slices; coordinator reconciles at integration |
| `docs/as-built/v504_quest-badge-ground-loot.md` | Slice evidence | Create only after implementation; coordinator owns final lifecycle/PROGRESS closeout |

Keep new source/test files below 600 lines. The transferred `loot_node_factory.gd` is 464 lines and `test_loot_node_factory.gd` is 256 lines; put shape geometry in the new helper so the factory stays below 600. Do not grow `client/scripts/main.gd` for this feature.

## Ordered tasks

Implementation starts only after the coordinator approves this revised plan. The user resolved key scope: existing drops only; keys are deferred. Reserve the shared Godot runner with the coordinator before any Godot, bot, or visual capture command.

### 1. Reconcile the v503 prerequisite and finalize the scope

- [x] Confirm detached `HEAD` remains the assigned SHA and the coordinator's 23 v503 paths are present as an uncommitted overlay; no transferred commit revision exists.
- [x] Inspect v503's as-built, item-presentation schema, resolver and pose path, transferred tests, and available manifest assets. Update this file map to the actual transferred state.
- [x] Record the user’s key decision in the spec: use current quest/trophy/badge-family drops and defer keys.
- [x] Obtain coordinator approval of this exact mapping, file map, and test coverage before implementation.

### 2. Add data-owned presentations for the accepted item set

Files: `shared/assets/item_presentations.v0.json` and schema, `client/scripts/loot_quest_badge_shapes.gd`, and a small dispatcher in `client/scripts/loot_node_factory.gd`.

- [x] Add the exact family/item shape mapping above. Retain existing badge-item palettes/scales; set each trophy's new ground palette and scale in its item override. Add only the six new `groundShape` enum values (`heart`, `wing`, `hooded_head`, `skull`, `shard`, `stone`).
- [x] Build the accepted mesh presets with low-poly materials. Keep overall scale and color in the shared catalog; shape geometry is renderer-owned, and no asset association or gameplay tuning is added. Real-renderer appearance still needs inspection.
- [x] Preserve the current GLB resolver and primitive fallback behavior for unrelated items. The chosen presets have focused nonzero-bounds assertions pending Godot execution; no manifest asset reference is added.
- [x] Existing `make validate-shared` validates the catalog/schema and CODEMAP (2,208 checks passed); no redundant validator is needed.

Smallest check: `make validate-shared`. This plan has no manifest or GLB change, so `make validate-assets` is optional unless implementation expands asset references after a new review.

### 3. Cover every accepted item and preserve the existing behavior

Files: `client/tests/test_loot_node_factory.gd`; only add wider presentation assertions if the focused factory test is insufficient.

- [x] Derive test item IDs from `items.v0.json` and resolved family data rather than copying a hand-maintained list.
- [x] Assert each accepted item instantiates its catalog-selected mesh composition, with a non-empty visual bound and its resolved ground preset from the shared catalog.
- [x] Assert quest and badge silhouettes differ at the item-model layer, and the four trophies have distinct shapes. Labels, rarity glow, pickup beam, and the fixed loot-root collider remain unchanged; existing tests cover gold amount.
- [x] Assert gold/potion fallback or models and hand-item/armor model paths remain unaffected, including the transferred v503 gear coverage.

Smallest check: `godot --headless --path client --script res://tests/test_loot_node_factory.gd`.

### 4. Add live client proof and inspect the real renderer

Files: the existing `quest_town_turn_in` and `blacksmith_renew_item` client scenarios; focused captures and evidence. The quest lab has a placed `quest_leaf`; the existing bishop debug path creates a `renew_stone` in the vendor lab. These are separate runs because the existing worlds do not place both families together. The existing protocol `steward_hunt_quest` scenario covers trophy pickup. No new world fixture or gameplay drop path is needed.

- [x] Add one `capture_frame` step before the existing `quest_leaf` pickup in `quest_town_turn_in`, and one after the bishop debug drop but before the existing `renew_stone` pickup in `blacksmith_renew_item`; mark both `skip_if_headless: true`. Keep their pickup assertions and scenario tiers unchanged.
- [x] Run both named headless client scenarios and `make bot scenario=steward_hunt_quest`; all passed.
- [x] Run both player-camera visual scenarios and preserve their captures. The blacksmith scenario passed with `BOT_STEP_DELAY=0.0`; its default delay timed out waiting for the preexisting bishop event before the new capture step.
- [x] Run `make regen-screenshots SUITE="floor-item"` (60/60) and focused real-renderer captures for quest, trophy, shard, stone, and badge items; inspect the focused results alongside the gear regression gallery.
- [x] Inspect and preserve the reviewed PNGs in `docs/as-built/assets/v504/`. Iterative first/final captures exist for leaf, hooded head, and skull; no pre-slice quest/badge baseline PNG was found. The live-camera leaf remains small in a dark lab, limiting a strong readability claim.

Smallest checks: the two named `make bot-client` runs above and the focused Godot/showme commands below.

### 5. Run focused gates and prepare handoff

- [x] Run `make client-unit`, `make validate-shared`, and `make maintainability`; no manifest/GLB reference changed, so `make validate-assets` was not required.
- [x] Run `git diff --check`; audit the shared catalog/schema changes and every file touched after v503 transfer.
- [x] Write the v504 as-built summary with focused command results, exact scenario/capture paths, and limits.
- [x] Report base and transferred revisions, complete changed/deleted/untracked file list, ignored evidence to preserve, overlap resolution, exact checks, and remaining acceptance gaps to the coordinator.

The batch coordinator runs the combined `make ci` only after all accepted slices are integrated. This slice does not run full CI, `/finish`, commit, or push.

## Verification commands

```bash
godot --headless --path client --script res://tests/test_loot_node_factory.gd
make client-unit
make validate-shared
make maintainability
make validate-assets  # only if a newly approved manifest/runtime asset reference changes
make bot-client SCENARIO=quest_town_turn_in HEADLESS=1
make bot-client SCENARIO=blacksmith_renew_item HEADLESS=1
make bot scenario=steward_hunt_quest
make bot-visual scenario=quest_town_turn_in
BOT_STEP_DELAY=0.0 make bot-visual scenario=blacksmith_renew_item
make regen-screenshots SUITE="floor-item"
python3 skills/showme/scripts/render_focus.py --focus floor-item --items quest_leaf
python3 skills/showme/scripts/render_focus.py --focus floor-item --items quest_trophy_bat_wing
python3 skills/showme/scripts/render_focus.py --focus floor-item --items respec_badge
git diff --check
```

**Review gate outcome:** The coordinator approved this plan and reserved the shared Godot runner before runtime work. The v503 gear-pose catalog and tests remain part of this worktree but are not v504-owned edits. The runner was released after focused verification; the coordinator owns batch integration, combined CI, and closeout.

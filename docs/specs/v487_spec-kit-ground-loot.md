# v487 — Kit ground loot: dropped weapons show the kit model the hero wields (ADR-0018 P3 follow-up)

- **Status:** Implemented (v487)
- **Date:** 2026-09-30
- **Codename:** `kit-ground-loot`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D5/D10, P3 row ("ground loot still legacy")
- **Baseline:** v486 + `$refactor` paydown (`01211039`)
- **Spec gate:** client presentation only (no protocol, server, rules or golden change). It
  qualifies for the CLAUDE.md spec-gate exemption; a spec and plan are written anyway.

## Purpose

Ground loot is the last shipped visual on the legacy pipeline, and it is visibly wrong:

- `LootNodeFactory.make_ground_equipment_model` resolves the model from the `3d_model` field of
  the item's `item_presentations.v0.json` **family**. Those fields still point at the pre-kit
  GLBs: `dagger`, `mace`, `hammer` and `spear` drops render as `weapon_rusty_sword_v0`, and
  `halberd` and `war_hammer` as `weapon_starter_axe_v0`.
- Meanwhile `shared/assets/item_visuals.v0.json`, which CLAUDE.md names the canonical
  item → asset link, already maps every weapon and off-hand to its KayKit model (v475). The hero
  wields that model, so the drop and the equipped item disagree.
- `LootNodeFactory.apply_model_tint` replaces every material with a flat untextured
  `StandardMaterial3D`, the exact bug `ModelTint` fixed for monsters in v474 (v486 client review).

After this slice:

1. A weapon or off-hand drop instantiates the **same asset `item_visuals.v0.json` assigns to the
   item** (the model in the hero's hand).
2. The ground pose (lie-flat rotation, height, scale) of kit models is presentation data in
   `shared/assets/equipment_display.v0.json`, not code literals.
3. The rarity tint of every ground model goes through `ModelTint`, so kit textures survive. Kit
   (rig-native) models blend the rarity colour at `rig_native_rarity_tint_strength`, the same
   strength equipped kit weapons use.
4. Armor and jewelry drops keep their family `3d_model` fallback GLB (KayKit armor is a body tint,
   not a mesh, per ADR-0018 D5). Weapon and off-hand families no longer carry a `3d_model`; the
   validator rejects one, so the two mappings cannot drift again.
5. `item_visuals.v0.json` is loaded through one static `ItemVisualsLoader` (agent rule 7), used by
   both `LootNodeFactory` and `EquipmentVisuals`, instead of adding another private JSON parse.

## Non-goals

- Vendoring new kit props (coins, potion bottles, keys from Dungeon Remastered) for gold,
  consumable or quest drops. They need licence, `gltf_to_glb`, manifest and budget work: a
  follow-up slice.
- Replacing the armor/jewelry fallback GLBs (no kit armor meshes exist).
- Purging the legacy weapon GLBs (`weapon_rusty_sword_v0` …). They are still referenced by
  `equipment_visuals.gd` fallbacks, the model viewer and asset catalog tests; purge is a separate
  cleanup once no runtime consumer remains.
- Unifying the rarity palettes (`equipment_visuals.gd` vs `loot_node_factory.gd` vs panels).
- Any gameplay, drop-rate, rarity, protocol or server change.

## Acceptance criteria

- [ ] For every `item_visuals.v0.json` entry with slot `main_hand` or `off_hand` whose item has a
  ground presentation, `make_loot_node` produces a child named `GroundModel_<item_visuals asset_id>`
  (derived from the catalog in the test, not pinned per item).
- [ ] Armor/jewelry drops (e.g. a `helm` family item) still produce their family `3d_model`
  fallback node.
- [ ] Items with no model (gold, potions, quest items) still render the primitive + rarity
  background, unchanged.
- [ ] A tinted kit ground model keeps its albedo texture (`ModelTint` duplicate, not a fresh
  material) and its `albedo_color` equals `WHITE.lerp(rarity_tint, rig_native_rarity_tint_strength)`.
- [ ] The kit ground pose is read from `equipment_display.v0.json` (`ground_pose`), schema-backed;
  a per-asset override is honoured. Tests derive the expected transform from the loaded data.
- [ ] `make validate-shared` fails if a weapon or off-hand family (a family whose items map to a
  `main_hand`/`off_hand` `item_visuals` entry) sets `3d_model`, and passes on the committed data.
- [ ] `EquipmentVisuals` and `LootNodeFactory` read `item_visuals.v0.json` only through
  `ItemVisualsLoader`.
- [ ] Visual gate (ADR-0018 D9): `make regen-screenshots SUITE="floor-item"` captures reviewed
  before/after; kit weapons lie flat on the ground, readable, not clipping into the floor. The
  as-built commits representative captures.
- [ ] `make client-unit` green; `make validate-shared` green.

## Scope and files likely touched

| Area | Files |
|------|-------|
| Shared presentation data | `shared/assets/item_presentations.v0.json` (drop weapon/off-hand family `3d_model`), `shared/assets/equipment_display.v0.json` + `.schema.json` (`ground_pose`) |
| Client | new `client/scripts/item_visuals_loader.gd`; `client/scripts/loot_node_factory.gd`; `client/scripts/equipment_visuals.gd` (use the loader); `client/scripts/equipment_display_loader.gd` (`ground_pose`) |
| Validation | `tools/validate_item_presentations.py` (+ its test if present) |
| Tests | `client/tests/test_item_visuals.gd`, `client/tests/test_loot_node_factory.gd`, a focused `test_item_visuals_loader.gd` if the loader is not covered elsewhere |
| Docs | ADR-0018 P3 row, `docs/CODEMAP.md` (new loader), `docs/as-built/v487_kit-ground-loot.md`, lifecycle row |

**Asset decision (adopt / borrow / reject):** *adopt* the already-vendored KayKit weapon/shield/
spellbook GLBs through the existing `item_visuals.v0.json` mapping; *reject* vendoring new props in
this slice (see non-goals); *keep* the generated armor/jewelry fallbacks as the explicit D10
fallback. No new plugin or asset.

## Test and bot proof

- Godot unit tests (catalog-derived, per the Test Locking Policy): ground model identity per
  item_visuals entry, armor fallback, primitive fallback, tint texture preservation and strength,
  pose from data.
- `make validate-shared` negative case for a weapon family `3d_model`.
- Visual regression: `make regen-screenshots SUITE="floor-item"` (human review, ADR-0018 D9).
- No new bot scenario: headless client bots run `gl_compatibility` and cannot judge the Forward+
  look (ADR-0018 D9); existing loot pickup scenarios keep proving the node still spawns and picks up.

## Open questions and risks

- **Kit weapon authoring orientation.** Kit weapons are authored for `handslot` bones (blade
  along +Y). The shared kit ground pose is tuned from captures; odd assets (bow, spellbook,
  shields) may need a per-asset override. Risk is visual only.
- **Pick collision.** The loot pick collider is built on the loot root, not the mesh; a different
  model size must not change pickup reach. Verified by the existing pickup scenarios.

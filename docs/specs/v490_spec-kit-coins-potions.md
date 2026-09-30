# v490 — Kit coins and potions: gold and potion drops use KayKit props (ADR-0018 D10 follow-up)

- **Status:** Implemented (v490)
- **Date:** 2026-09-30
- **Codename:** `kit-coins-potions`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D10 (procedural shapes are test-only/explicit fallbacks)
- **Baseline:** v489 (`b3d32d23`)
- **Spec gate:** client presentation + presentation data only; exempt, written anyway.

## Purpose

After v487, weapons on the ground are kit models, but gold and potions are still `add_loot_primitive`
cylinders and boxes. After this slice:

1. **Gold** drops pick a KayKit Dungeon Remastered coin model by amount: `coin`, then
   `coin_stack_small`, `coin_stack_medium` and `coin_stack_large` from data tiers
   (`item_presentations` gold family `ground_model_tiers`: `min_amount` → `asset_id`).
2. **Potions** (health, mana, rejuvenation families) use kit bottles through the existing family
   `3d_model` fallback: `bottle_A_labeled_brown`, `bottle_B_brown`, `bottle_C_brown`.
3. A new optional family `ground_tint` (`color`, `strength`) colours the model through the material
   **detail layer**: it lerps the atlas toward the colour and keeps its shading, where a multiply
   would darken a brown bottle into mud. Families with `ground_tint` skip the flat rarity tint.
4. The detail-layer primitives move out of `ArmorLook` into a shared `ModelDetailTint` module
   (`set_detail`, `clear_detail`), which ArmorLook and the loot factory both use.
5. Coins and bottles stand upright via `equipment_display.v0.json` `ground_pose.assets` overrides
   (v487 pose contract: rotation, `rest_on_floor`, `max_extent`).

## Non-goals

- Quest items, keys, badges and other primitive loot shapes.
- Rarity visuals for gold/potions, pickup/label changes, any drop or gameplay change.
- Further ArmorLook refactors beyond moving the detail helpers.

## Acceptance criteria

- [ ] Gold drops choose the highest tier whose `min_amount` ≤ amount (derived from the catalog in the
  test); an amount below the first tier keeps the primitive.
- [ ] Each potion family's drop instantiates its `3d_model` bottle; its mesh has the detail layer
  enabled with the family `ground_tint` colour/strength and keeps the albedo texture.
- [ ] Weapon drops are unchanged (v487 tests green); ArmorLook tests green after the extraction.
- [ ] `make validate-shared` accepts the new schema fields and rejects a tier with an unknown asset;
  `make validate-assets` accepts the six new CC0 manifest entries within D8 budgets.
- [ ] Visual gate: floor captures of gold amounts and the three potions, before/after in the as-built.
- [ ] `make client-unit` green; loot pickup client scenario passes.

## Files

`assets/manifests/assets.v0.json` + `client/assets/environment/kaykit_dungeon/{coin,coin_stack_small,coin_stack_medium,coin_stack_large,bottle_A_labeled_brown,bottle_B_brown,bottle_C_brown}.glb`;
`shared/assets/item_presentations.v0.json` + schema; `shared/assets/equipment_display.v0.json`;
`tools/validate_item_presentations.py` (+ test); new `client/scripts/model_detail_tint.gd`;
`client/scripts/armor_look.gd`; `client/scripts/loot_node_factory.gd`; tests; `skills/showme` floor-item capture of gold/potions if needed; docs.

**Asset decision:** adopt seven CC0 KayKit Dungeon Remastered 1.0 props (already staged); no plugin.

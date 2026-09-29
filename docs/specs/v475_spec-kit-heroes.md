# v475 — KayKit heroes, Rig_Medium clips, rig-native weapons (ADR-0018 P3a)

- **Status:** Implemented (v475)
- **Date:** 2026-09-29
- **Codename:** `kit-heroes`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D4/D5, phase P3
- **Baseline:** `c9b5f4d7`
- **Inventory:** [kaykit-asset-inventory.md](../researchs/kaykit-asset-inventory.md)

## Purpose

The five hero classes switch to **KayKit Adventurers 2.0** models, animated by **Character
Animations 1.1 `Rig_Medium`** clips, and hold rig-native weapons.

| Class | Model |
|-------|-------|
| paladin | Knight |
| barbarian | Barbarian |
| sorcerer | Mage |
| rogue | Rogue |
| ranger | Ranger |

1. **Clip library.** The new `KitHeroClips` builds one cached `AnimationLibrary` per catalog profile
   from the vendored `Rig_Medium_*` clip GLBs, under the **logical** names the game already plays:
   - `idle`, `walk`, `hit`, `death`
   - `attack`, `attack_off_hand`, `attack_2h`, `attack_ranged`, `attack_staff`

   Idle and walk are looping copies. `character_visual.gd` swaps this library into the hero's
   `AnimationPlayer` whenever the model changes (`refresh_gear_sockets()`, which every model-swap
   site already calls). The legacy `character_anims.tres` is restored for non-kit models. No
   gameplay, event or clip-name changes.
2. **Sockets.**
   - `gear_sockets.v0.json` defaults move to kit bones: `handslot.r/l`, `head`, `chest`, `hand.r/l`,
     `hips`, `foot.l/r`.
   - An optional `fallback_bone` keeps the legacy 17-bone fallback model (`base_human`, used for
     unknown classes and hero corpses) working until P3c deletes it.
   - Socket names are unchanged.
3. **Rig-native weapons.**
   - **Kit accessories** (authored for `handslot`, identity transform): Adventurers 2.0 swords, axes,
     dagger, bow, staff, wand, shields, spellbook.
   - **FantasyWeaponsBits** for the families Adventurers lacks: hammer, mace, spear, halberd.
     As built, these share the kit convention (origin at the grip, +Y toward the business end),
     so they also mount at **identity**. The spec originally assumed they needed tuned transforms.
   - `item_visuals` main-hand and off-hand entries are remapped per family.
   - A new `rig_native: true` flag on those entries skips the legacy equipped-scale multiplier and
     the `ModelRoot` scale compensation, which assumed world-meter weapons.
   - Equipped-item rarity tint uses `ModelTint`, so kit weapon textures survive.
4. **Manifest.**
   - Heroes are `character` assets. `required_nodes` holds the socket bones.
   - Clip GLBs use a new `type: animation`, with its own D8 budget; it covers the preview mannequin
     mesh the library files carry.
   - The weapons are `equipment` assets with slots.
   - Validator step [5] now derives the required hand bones from `gear_sockets`
     (`bone` or `fallback_bone`) instead of hardcoding `hand_r`/`hand_l`.

## Non-goals

- **P3b:** armor as tints (chest, gloves, boots, belt) and headgear toggles. In v475 the legacy
  procedural armor boxes still mount on the kit sockets.
- **P3c:** deleting the legacy pipeline — `base_human`, class GLBs, the morph and rig generators,
  `character_anims.tres` and its tests.
- FantasyWeaponsBits variety beyond the four missing families.
- First-person camera retuning beyond keeping `chest_view` usable.

## Acceptance criteria

1. **Clips.** Each class resolves a kit model. Its `AnimationPlayer` holds every logical clip the
   game plays; idle and walk loop and death does not. The legacy model keeps
   `character_anims.tres`.
2. **Sockets.** Right-hand, off-hand, head and chest sockets bind to the kit bones on kit heroes,
   and to the legacy bones on `base_human`.
3. **Weapons.** Every `item_visuals` main-hand and off-hand entry resolves to a manifest `equipment`
   asset in the right slot. Rig-native weapons mount under their socket at identity scale, times the
   entry scale. Rarity tint keeps the albedo texture.
4. **Validation.** `validate-assets` passes, with the D8 budgets for `character`, `equipment` and
   `animation`, and the gear-socket-derived hand-bone rule.
5. **Visuals.** The showme `gear` and `classes` captures show textured kit heroes holding weapons
   correctly. The as-built compares them to the v469 baseline.
6. **Gates.** `make ci` passes. Tests that pinned the 17-bone rig move to catalog- and
   socket-derived checks.

## Asset/plugin decision

- **Adopt:** KayKit Adventurers 2.0, Character Animations 1.1 and Fantasy Weapons Bits 1.0 (CC0,
  staged per the inventory).
- **Reuse:** the v474 `gltf_to_glb` packer, the `ModelTint` helper and the `KitPieceLibrary`
  manifest resolver.
- **Reject:** retargeting (not needed: same deform bones), and Godot's `AnimationLibrary` importer
  mode (it would need per-file import settings; instancing the clip GLB once and caching is
  simpler).

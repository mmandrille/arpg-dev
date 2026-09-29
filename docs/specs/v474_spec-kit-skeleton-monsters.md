# v474 — KayKit skeleton monsters and monster asset purge (ADR-0018 P4a)

- **Status:** Implemented (v474)
- **Date:** 2026-09-29
- **Codename:** `kit-skeleton-monsters`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D1/D2/D4/D8, phase P4
- **Baseline:** v473 (`bf4ba438`)

## Purpose

1. **Fix: monster textures never render.** `main._apply_model_tint` replaces every monster mesh
   material with a plain `StandardMaterial3D` of the rarity tint. That throws away the texture,
   which is why every GLB monster renders flat white or tinted (the v469 finding). The tint must
   multiply a **copy** of the mesh's own material, so the texture survives. `ModelReactionController`
   already duplicates `material_override`, so hit flashes keep working.
2. **Humanoid monsters become KayKit Skeletons 1.0** (CC0; same 41-joint rig as the Adventurers,
   95 embedded clips):

   | Monster def | Kit model | Weapons (on handslot bones) |
   |-------------|-----------|-----------------------------|
   | `dungeon_mob` | Skeleton_Warrior | Skeleton_Axe + Skeleton_Shield_Small_A |
   | `dungeon_archer` | Skeleton_Rogue | Skeleton_Crossbow |
   | `dungeon_undead` | Skeleton_Minion | Skeleton_Blade |

   Weapons are attached with `BoneAttachment3D` on `handslot.r` / `handslot.l`.
3. **Clip catalog (ADR-0018 D4 starter).** A new `shared/assets/kit_monster_presentation.v0.json`
   holds:
   - `clip_profiles`, which alias the logical clips `AnimationController` plays (`idle`, `walk`,
     `attack`, `hit`, `death`, `dive`, `pounce`) to kit clip names
   - per-scene entries: asset, clip profile, attachments

   A scene-root script (`KitMonsterVisual`) adds the aliases to the imported `AnimationPlayer` in
   `_ready`. That runs before main builds the `AnimationController`, because the node is added to
   the tree first.
4. **Purge (ADR-0018 D2).**
   - Remove `dark_purple_monster` and `crocodile_archer` everywhere:
     - their license is unconfirmed, and dark_purple is over budget
     - runtime GLBs, sources, scenes, `.tres` libraries, manifest entries, the budget exemption
     - schema enum values and preview-catalog entries
     - `tools/assets/rig_monster_glbs.py`, which only rigged these two, plus its test and its
       `gen-assets` line
   - Delete the three unused source GLBs: `rock_demon`, `goblin_warrior`, `shadow_spirit`
     (≈17.9 MB, referenced by nothing).

## Non-goals (P4b)

Beasts keep their current unconfirmed-license models: `dungeon_wolf` (fox quadruped),
`companion_black_wolf` (wolf) and `dungeon_bat` (tiny flyer). KayKit has no beasts. P4b needs a
style-matched CC0 source (for example Quaternius) and its own rigs and animation profiles.

The following also stay unchanged: bosses (their data doesn't use the purged models), training
dummies, mercenaries and heroes (P3).

## Acceptance criteria

1. **Tint preserves textures.** For a textured mesh, `ModelTint.tinted_material(mesh, color)`
   returns a material that keeps the source `albedo_texture` and sets `albedo_color` = tint.
   Untextured meshes behave as before.
2. **New assets.** 3 kit skeleton `.glb` files plus 4 weapon GLBs (converted from `.gltf`/`.bin`/PNG
   by the new deterministic `tools/assets/gltf_to_glb.py`, with its own tests). Every one:
   - is a `type: monster` manifest asset with provenance and license. The weapons are monster
     accessories, not player equipment, so they have no slot and no `item_visuals` link.
   - is within the D8 budgets
   - skeletons declare `required_nodes` = the kit bones that attachments use
3. **Validation.** `kit_monster_presentation.v0.json` plus its schema validate, and every asset id in
   it resolves in the manifest (a validator cross-check).
4. **Kit scenes.** Each kit monster scene has a model, an `AnimationPlayer` with every logical clip in
   its profile, and weapons on the named bones.
5. **Mapping.** `monster_visuals.v0.json` maps the three defs to the kit scenes. The `family_accent`
   and scale follow the catalog.
6. **Purge is complete.** Nothing references `dark_purple` or `crocodile` except history docs. The
   budget file has no exemption for them. `make validate-assets` shows no orphans.
7. **Visuals.** The `monsters` showme focus shows textured kit skeletons with weapons, and the
   as-built compares it to the v469 baseline.
8. **Gates.** `make ci` passes. The client bot scenarios that fight `dungeon_mob`, `dungeon_archer`
   or `dungeon_undead` still pass, because gameplay is unchanged.

## Asset/plugin decision

- **Adopt** KayKit Skeletons 1.0 (staged and verified in v469).
- **Borrow** the v469 `glb_reader` for the converter.
- **Reject** shipping `.gltf` + `.bin`, which breaks the manifest single-file contract (ADR-0006 D1),
  and reject retargeting the old monsters.

## Risks

- **Repository size.** The embedded 95-clip libraries make each skeleton about 4.8 MB, roughly
  +14.5 MB. The purge removes about 36 MB, so the net change is a decrease.
- **Monsters with `visual_model` overrides** (bosses) are untouched.

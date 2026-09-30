# v484 — Retire the legacy 17-bone hero pipeline (ADR-0018 P3c)

- **Status:** Implemented (v484)
- **Date:** 2026-09-29
- **Codename:** `retire-legacy-hero`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D4 ("Retired" list), phase P3c
- **Baseline:** v483 (`5e891e5e`)

## Purpose

v475 moved the five classes to KayKit Adventurers heroes, but the legacy pipeline is still wired
in. A research pass found it is not just dead weight:

- **Every remote co-op player renders as the grey `base_human` mannequin.** The server never sends
  `character_class` on player entities, so `ClassPresentationsLoader.resolve("")` falls back to
  `base_human` and its 17-bone clips.
- The **hero corpse**, the boss `current_humanoid_player` model, the model viewer, `smoke.gd` and
  the class-less showme captures also instance the bare `character.tscn`, which embeds `base_human`
  and `character_anims.tres`.
- The model-viewer test only passes because the legacy library's clip *names* exist; its tracks
  don't bind to the kit rig.
- `ClassBodyTint` (ADR D4: retired) still multiplies every kit texture toward a class colour. Its
  skin colour is also the reaction base tint that remote-player status tints flatten meshes to,
  which turns textured kit heroes orange.

## Scope

1. **Data-driven kit fallback.** `class_presentations.v0.json` gains `fallback_class` (`paladin`).
   `resolve()` of an empty or unknown class returns that class's kit model and `kaykit_hero` clips.
   `character.tscn` embeds the fallback kit hero with an empty `AnimationPlayer`, and
   `character_visual._ready` syncs the kit clip library. So every bare instance shows an animated
   kit hero: remote players until the server sends their class (follow-up task), the boss humanoid
   model, smoke, and showme.
2. **Hero corpse** (`hero_corpse_visual.gd`, extracted so `main.gd` doesn't grow) uses the fallback
   kit hero, frozen at the end of its authored `death` clip, instead of rotating the mannequin
   -88°. The gold corpse tint and the pick box are unchanged.
3. **Retire `ClassBodyTint`**:
   - delete the script, its test and smoke gate, and `body_tint` from the data, schema and
     validator;
   - reaction and remote-player base tints become white, so kit textures render as authored.
4. **Remove the 17-bone socket fallbacks.** Delete the `fallback` blocks from `gear_sockets.v0.json`
   (and the schema), `character_visual.gd`, the test helpers, and the `hand_r`/`hand_l` default in
   `validate_assets.py`.
5. **Fix hidden legacy dependencies**:
   - the model viewer loads kit clips;
   - the showme skeleton focus uses kit bone names;
   - `inspect_rig.gd` stops checking `base_human`.
6. **Delete the pipeline:**
   - `base_human` and the five unused legacy class GLBs, with their manifest entries and sources;
   - `character_anims.tres` and the character clips in `build_animations.gd`;
   - `class_body_morph.py`, `rig_hero_glbs.py`, `rig_canonical_hero.py`, `canonical_skeleton.py`,
     `fix_skeleton.py`, `base_human_mesh.py`, `skin_blend.py`, the humanoid builders in
     `gen_glb.py`, and their tests;
   - the `gen-assets` lines.

   The GLB helpers that `import_equipment_glb.py` still imports from `rig_hero_glbs.py` move to
   `tools/assets/glb_mesh_io.py` first.

## Non-goals

- **Server sends `character_class` on player entities.** This is a separate protocol slice, filed as
  a follow-up. The client already routes a class to `_apply_character_class_model` once it arrives.
- **Class-specific corpses.** There is no class field on corpse interactables; adding one needs a
  protocol change, server, store, migration and replay snapshot.
- The first-person polish, ground loot and `monster_anims.tres` (the monster dummy and skeleton keep
  it).

## Acceptance criteria

1. **Bare scene.** `character.tscn` instanced with no class shows the fallback kit hero, and the
   `AnimationPlayer` plays `idle`/`walk`/`hit`/`death` kit clips bound to the kit skeleton.
2. **Unknown class.** `resolve("")` and `resolve("nope")` return the `fallback_class` model with
   `clip_profile: kaykit_hero`. No code path references `character_base_human_v0`.
3. **Hero corpse.** It renders the kit hero lying down through the `death` clip, with the gold tint,
   the shadow and the label. The co-op corpse tests pass.
4. **No body tint.** No `ClassBodyTint` or `body_tint` remains. The local and remote reaction base
   tint is white.
5. **Sockets.** The gear sockets have no `fallback`. `validate_assets` requires kit bones.
6. **Deletion.** The deleted files have no remaining references (codemap, make targets, the
   file-size baseline). `make validate-assets` reports no orphans.
7. **CI.** `make ci` passes, and the showme `corpse`, `gear` (no class), `companions` and
   `skeleton` captures show kit heroes.

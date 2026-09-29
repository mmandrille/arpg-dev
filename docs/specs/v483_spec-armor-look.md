# v483 — Armor look: tints on the kit hero's own body parts, headgear toggle (ADR-0018 P3b)

- **Status:** Implemented (v483)
- **Date:** 2026-09-29
- **Codename:** `armor-look`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D5, phase P3b
- **Baseline:** v482 (`b0569096`)

## Purpose

v475 moved heroes to KayKit, but equipped armor still mounts the legacy procedural GLB boxes (helm,
mail, gloves, belt, boots, ring, amulet) on kit sockets, for example the gold lump on the shoulder.
It is the most visible defect on screen: the player looks at their own hero all the time.
ADR-0018 D5 decided the replacement:

| Slot | World visual (this slice) |
|------|---------------------------|
| chest | tint of the hero's `*_Body` mesh |
| belt | tint of `*_Body` **only when no chest item is equipped** (chest wins; D5 left the precedence to P3) |
| gloves | tint of `*_ArmLeft` + `*_ArmRight` (whole arms) |
| boots | tint of `*_LegLeft` + `*_LegRight` (whole legs) |
| head | the class's own headgear meshes (`*_Helmet`, `*_HelmetVisor`, `*_BearHat`, `*_Hat`), **shown only while a head item is equipped** and tinted by that item |
| ring_left, ring_right, amulet | no world visual (UI only) |
| main_hand, off_hand | unchanged (v475 rig-native kit weapons) |

## Design

**Catalog `shared/assets/armor_look.v0.json`** (schema-backed presentation data):
- `regions`: region id → mesh-name suffixes. Suffix matching works for every kit hero
  (`Knight_Body`, `Mage_Hat`, …), so no per-class table is needed.
- `slots`: slot → `{region, mode: "tint" | "headgear"}`.
- `body_precedence: ["chest", "belt"]`.
- `no_world_visual_slots`: the rings and the amulet.
- `default_strength`.
- `items`: item_def_id → `{color, strength?}`. Item templates carry no colour, so every tint-slot
  template gets a material colour here (leather brown, steel grey, robes purple, …).

**Mechanism: the material detail layer, not `albedo_color`.**
- The kit atlases are already coloured, so an `albedo_color` multiply can only darken them. Brown
  leather over a blue tunic comes out muddy blue.
- Instead, `StandardMaterial3D.detail_albedo` is set to a 1×1 texture with the colour and
  alpha = strength, and the detail blend mode is Mix. That lerps the texture toward the colour. The
  probe gave a readable brown tunic that keeps its belt and emblem detail.
- `albedo_color` stays owned by the class body tint, the rarity tint and `ModelReactionController`
  (hit flash, death darken). The reaction controller writes `albedo_color` on the **current**
  `material_override`, and the armor look edits that same material in place, so neither erases the
  other. This is the risk the research flagged.
- Applying is idempotent. Each refresh recomputes every region from the equipped set and clears
  regions with no item.

**Code**
- New `client/scripts/armor_look.gd`: loader plus presenter, using the `class_name` static pattern
  (Agent rule 7).
- `EquipmentVisualResolver` stops mounting meshes for tint, headgear and no-visual slots. It
  applies the armor look on the mount root's `ModelRoot` and reports
  `kind: "tint" | "headgear" | "none"` in its debug state.
- The boots mirror mount and the armor procedural fallbacks become dead code and are removed.

## Non-goals

- **Remote players and mercenaries.** They don't show gear today, and the snapshot protocol carries
  no equipment for other entities, so this needs a protocol slice.
- **Ground loot** keeps the legacy `item_presentations` `3d_model` GLBs. The fallback armor manifest
  entries and GLBs stay because loot still uses them.
- **Rogue and Ranger have no headgear mesh.** Their head items show nothing (a documented gap). The
  only standalone kit headgear is the Skeletons-pack helmets, which are sized for skeleton heads.
- **Rarity is not expressed on armor tints**; loot labels and icons already carry it.
- No gameplay, rules, protocol or golden changes.

## Acceptance criteria

1. **Catalog coverage.** The catalog and schema validate. Every template in a tint or headgear slot
   has an `items` colour. Every equippable template is covered either by `item_visuals` (weapons)
   or by `armor_look` (a Python gate plus a client test).
2. **Tint slots.** Equipping chest, gloves or boots changes the detail tint of the matching part
   meshes. Unequipping clears it. No mesh is mounted under the armor sockets.
3. **Belt precedence.** With a chest item equipped, the belt does not change the body. With no chest
   item, the belt tints the body.
4. **Headgear.** A head item shows the class headgear with its tint, and an empty head slot hides
   it. Rogue and Ranger degrade to no visual without warnings.
5. **Hit flash.** A hit flash after equipping keeps the tint, and a later equip change keeps the
   reaction working.
6. **Visuals.** The showme `gear` focus shows tinted heroes for all five classes, with a
   before/after in the as-built.
7. **CI.** `make ci` passes.

## Asset / plugin decision

- **Adopt** kit body-part meshes plus the StandardMaterial3D detail layer. No shader and no new
  assets.
- **Reject** the `albedo_color` multiply, because it can't recolour a coloured atlas.
- **Reject** region-mask shaders, which ADR-0018 already ruled out.
- **Reject** new armor meshes, per ADR-0018 D5.

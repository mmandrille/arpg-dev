# KayKit asset inventory (staged packs, 2026-09-29)

This is the single reference for which KayKit packs are staged locally, what each contains, what
it is for in ADR-0018, and the format and rig facts later slices need.

- **Staging** (gitignored, ADR-0018 D2): `.artifacts/kaykit/`.
- **Rebuild a report:** `make inspect-kit KIT=<dir> OUT=<dir>`.
- **Existing reports:** `.artifacts/kaykit/report/itch/<pack>/report.{json,md}`.
- **Related:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) ·
  [v469 P0 findings](v469_kaykit-p0-findings.md).

## Staged archives

Every pack's `License.txt` states **CC0 1.0**, so all of them are on the D2 allow-list and need no
attribution.

| Pack | Version | Source | Archive (under `.artifacts/kaykit/`) | sha256 | Extracted to |
|------|---------|--------|---------------------------------------|--------|--------------|
| Dungeon Remastered | 1.0 | GitHub `KayKit-Game-Assets` @ `b0ca9bd96a80` | `archives/KayKit-Dungeon-Remastered-1.0-b0ca9bd96a80.zip` | `e96a65ce…f714b` | `extracted/KayKit-Dungeon-Remastered-1.0-*` |
| Adventurers | 1.0 | GitHub @ `672074b73ba2` | `archives/KayKit-Character-Pack-Adventures-1.0-672074b73ba2.zip` | `00777a7b…b47af` | `extracted/KayKit-Character-Pack-Adventures-1.0-*` |
| Skeletons (rigged) | 1.0 | GitHub @ `15b62b9bad12` | `archives/KayKit-Character-Pack-Skeletons-1.0-15b62b9bad12.zip` | `2e3e1107…b8e6f` | `extracted/KayKit-Character-Pack-Skeletons-1.0-*` |
| **Adventurers** | **2.0 (free)** | itch.io, owner download | `archives/itch/adventurers_2.0.zip` | `abe48f47…dc479` | `itch/adventurers_2.0/` |
| **Character Animations** | **1.1** | itch.io, owner download | `archives/itch/character_animations_1.1.zip` | `65882f31…d20f` | `itch/character_animations_1.1/` |
| Dungeon Pack | 1.1 (free) | itch.io, owner download | `archives/itch/dungeon_pack_1.1.zip` | `6acb859d…d6a1c` | `itch/dungeon_pack_1.1/` |
| Fantasy Weapons Bits | 1.0 (free) | itch.io, owner download | `archives/itch/fantasy_weapons_bits_1.0.zip` | `1017144f…2d2` | `itch/fantasy_weapons_bits_1.0/` |
| Resource Bits | 1.0 (free) | itch.io, owner download | `archives/itch/resource_bits_1.0.zip` | `7056f131…cfcf7` | `itch/resource_bits_1.0/` |
| Skeletons (legacy) | 1.0 | itch.io, owner download | `archives/itch/skeletons_1.0.zip` | `a26d454a…a89b1` | `itch/skeletons_1.0/` |

Full sha256 values are in `shasum -a 256 .artifacts/kaykit/archives/**/*.zip`.

## Rig compatibility (resolves the ADR-0018 P3 open item)

| Rig | Joints | Where |
|-----|--------|-------|
| KayKit 1.0 rig | 41: 23 deform joints plus 18 IK/control helpers (`IK-*`, `control-*`, `*IK.*`) | Adventurers 1.0, rigged Skeletons 1.0; **in use** by the v474 kit monsters |
| **`Rig_Medium`** | **23: exactly the 1.0 deform joints, same names** | Adventurers 2.0 characters; Character Animations 1.1 `Rig_Medium_*` |
| `Rig_Large` | 23, separate proportions | Character Animations 1.1 `Rig_Large_*` (for the paid Barbarian_Large; unused) |
| Mannequin_Medium | 21 | Preview mannequin only |

The deform joints are `root`, `hips`, `spine`, `chest`, `head`, and per side (`.l` / `.r`):
`upperarm`, `lowerarm`, `wrist`, `hand`, `handslot`, `upperleg`, `lowerleg`, `foot`, `toes`.

**What this means:**
- Godot animation tracks are keyed by bone name, so **Character Animations 1.1 `Rig_Medium` clips
  drive the Adventurers 2.0 heroes directly**, with no retargeting.
- The same clips can also drive the v474 1.0-rig skeleton monsters. The helper bones just go
  unanimated.
- Adventurers 2.0 characters **embed no clips**. All hero animation comes from the 1.1 libraries.

## Pack contents and intended use

### Adventurers 2.0: heroes (ADR-0018 P3)

| Class | Model | Triangles | Height | Mesh parts (tint regions, ADR-0018 D5) |
|-------|-------|-----------|--------|----------------------------------------|
| paladin | `Knight.glb` | 5.8k | 2.54 | `Knight_{Body,ArmLeft,ArmRight,LegLeft,LegRight,Head}` + `Knight_Helmet`, `Knight_HelmetVisor`, `Knight_Cape` |
| barbarian | `Barbarian.glb` | 7.1k | 2.40 | `Barbarian_*` + `Barbarian_BearHat` |
| sorcerer | `Mage.glb` | 6.7k | 2.65 | `Mage_*` + `Mage_Hat`, `Mage_Cape` |
| rogue | `Rogue.glb` | 7.6k | 2.18 | `Rogue_*` + `Rogue_Cape` |
| ranger | `Ranger.glb` | 8.9k | 2.27 | `Ranger_*` + `Ranger_Cape`, `Ranger_Quiver` |
| (spare) | `Rogue_Hooded.glb` | 7.2k | 2.17 | `RogueHooded_*` (**no underscore after `Rogue`**) + `Mask` |

All six characters:
- are **skinned to `Rig_Medium`**, with `handslot.l` / `handslot.r` bones
- use one 1024² atlas each
- fit the ≤ 10k D8 budget

**Weapons are separate `.gltf` files** (31 accessories). They need `tools/assets/gltf_to_glb.py`
before vendoring:
`sword_1handed`, `sword_2handed`, `axe_1handed`, `axe_2handed`, `dagger`, `bow`,
`bow_withString`, `crossbow_*`, `staff`, `wand`, `spellbook_*`, `shield_*` (+ `_color`),
`quiver`, and arrows.

### Character Animations 1.1: hero clip libraries (P3)

`Rig_Medium_*.glb` (clips only):

| Library | Clips | Notable |
|---------|-------|---------|
| General | 15 | `Idle_A/B`, `Hit_A/B`, `Death_A/B` (+ `_Pose`), `Interact`, `PickUp`, `Throw`, `Use_Item`, `Spawn_*` |
| MovementBasic | 11 | `Walking_A/B/C`, `Running_A/B`, `Jump_*` |
| MovementAdvanced | 13 | `Dodge_*`, `Running_HoldingBow`, `Sneaking`, `Crouching`, strafes |
| CombatMelee | 22 | `Melee_1H_Attack_*`, `Melee_2H_Attack_*`, `Melee_Dualwield_Attack_*`, `Melee_Block*`, `Melee_Unarmed_*` |
| CombatRanged | 20 | `Ranged_Bow_Draw/Release(_Up)`, `Ranged_1H/2H_Shoot`, `Ranged_Magic_Shoot/Raise/Summon/Spellcasting` |
| Special | 15 | `Skeletons_*`: walking, death, resurrect, awaken, taunt, spawn |
| Simulation / Tools | 14 / 29 | Emotes, sitting; chopping, mining, fishing (town life) |

**Proposed hero logical-clip map (P3):**

| Logical clip | Kit clip |
|--------------|----------|
| idle | `Idle_A` |
| walk | `Walking_A` |
| attack | `Melee_1H_Attack_Chop` |
| attack_off_hand | `Melee_Dualwield_Attack_Stab` |
| attack_2h | `Melee_2H_Attack_Chop` |
| attack_ranged | `Ranged_Bow_Release` (bow) / `Ranged_2H_Shoot` (crossbow) |
| attack_staff | `Ranged_Magic_Shoot` |
| hit | `Hit_A` |
| death | `Death_A` |

This closes the v469 bow gap.

### Fantasy Weapons Bits 1.0: item visuals (P3)

It has 31 weapons (`.gltf` + shared `weapons_bits_texture.png`), all within the D8 equipment budget
(≤ 2.4k triangles). Mapping to `item_presentations` families:

| Family | Candidate models |
|--------|------------------|
| sword | `sword_A`, `sword_B` |
| greatsword | `sword_D`, `sword_E` |
| dagger | `dagger_A`, `dagger_B` |
| axe | `axe_A`, `axe_C` (1H), `axe_B` |
| hammer / war_hammer / mace | `hammer_A`, `hammer_B`, `hammer_C` |
| spear | `spear_A` |
| halberd | `halberd` |
| bow | `bow_A_withString`, `bow_B_withString` |
| staff | `staff_A`, `staff_B` |
| wand | `wand_A` |
| shield | `shield_A`, `shield_B`, `shield_C` |

Also included: `fistweapon_*`, `arrow_*`.

**Gaps:** no mace-specific model (use a hammer) and no book (use the Adventurers `spellbook_*`).

### Resource Bits 1.0: materials and loot (future)

76 models: copper, iron, silver and gold nuggets and bars; stone; wood; textiles; fuel barrels;
parts. Candidate uses:
- `upgrade_shard` (the only material currency today): silver or iron nugget visuals
- ADR-0012 "advanced dungeon resources" when they land

Budget watch: the `*_Stack_Large` / `*_Barrels` / `Wood_Log_Stack` pieces are 3.5–5.2k triangles,
over the D8 4k environment budget. Use single pieces.

### Dungeon Pack 1.1: superset of Dungeon Remastered 1.0

Every piece v471/v473 vendored (`wall`, `wall_half`, `pillar`, the floor tiles, `torch_mounted`,
`chest`, `chest_gold`) is **geometrically identical** in 1.1: same triangle counts and extents.
Keep the vendored 1.0 files. 1.1 adds these, which matter for the v473 gaps:
- `ceiling_tile`: a candidate for the procedural dungeon ceiling
- `stairs_modular_{left,center,right}` and `stairs_long*`: the modular pieces may fit the stair
  interactable footprint better than the 5 m `stairs`. This needs measuring.

### Skeletons 1.0 (itch archive): legacy, do not use

This is an **older, different pack** (`character_skeleton_*`, plus `_broken` variants):
- static **unskinned** meshes with no rig and no clips
- no embedded texture
- the models differ from the rigged GitHub Skeletons 1.0 that v474 uses

Keep it only for its possible `bone`, `skull` and `body_broken` debris props (corpse / ambient
dressing).

## Format notes

- **Characters and dungeon pieces ship as `.glb`**, and can be vendored as-is.
- **Accessories ship as `.gltf` + `.bin` + a shared texture PNG** (weapons, resource bits,
  Adventurers accessories). Pack them with `python -m tools.assets.gltf_to_glb <in.gltf> <out.glb>`
  (deterministic) before vendoring, per ADR-0006 D1.
- FBX, OBJ, DAE and `fbx(unity)` copies are ignored.
- The `Samples/` folders hold preview PNGs only, including art of the paid-tier Druid, Engineer and
  Barbarian_Large.

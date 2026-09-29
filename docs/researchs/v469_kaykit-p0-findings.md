# v469 — KayKit P0 findings (ADR-0018 verification checklist)

- **Date:** 2026-09-28
- **Slice:** [v469 spec](../specs/v469_spec-art-baseline-and-kit-verification.md) ·
  [plan](../plans/v469_2026-09-28-art-baseline-kit-verification.md)
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md)
- **Evidence (local, gitignored):**
  - `.artifacts/kaykit/report/report.{json,md}` (`make inspect-kit KIT=.artifacts/kaykit/extracted`)
  - `.artifacts/wall-grid-audit.json` (`make wall-grid-audit EXTRA_TILES=4`)
  - `.artifacts/kaykit/probe/`
- **Committed evidence:** [`docs/as-built/assets/v469/`](../as-built/assets/v469/)

## Sources staged

All three packs were downloaded directly from the author's official GitHub account, pinned by
commit, and extracted to `.artifacts/kaykit/extracted/`.

| Pack | Version | Source (commit) | Archive sha256 | License evidence |
|------|---------|-----------------|----------------|------------------|
| Dungeon Remastered | 1.0 | `github.com/KayKit-Game-Assets/KayKit-Dungeon-Remastered-1.0` @ `b0ca9bd96a80` | `e96a65ce4040b6b630f04470a60fd523d621c124cfbad062809cf7b94ccf714b` | `LICENSE.txt`: "License: (Creative Commons Zero, CC0)" |
| Character Pack: Adventurers | 1.0 | `github.com/KayKit-Game-Assets/KayKit-Character-Pack-Adventures-1.0` @ `672074b73ba2` | `00777a7b9811a7c7f8dd3fda4e56604f6a93b33673d64557153f354d677b47af` | `LICENSE.txt`: CC0 |
| Character Pack: Skeletons | 1.0 | `github.com/KayKit-Game-Assets/KayKit-Character-Pack-Skeletons-1.0` @ `15b62b9bad12` | `2e3e1107009fcb1ff4b2bccfe91c7e823264e267f20124438916f21a2d6b8e6f` | `LICENSE.txt`: CC0 |

**Not staged. These need a manual owner download**, because itch.io serves free packs through a
"name your price" dialog (ADR-0018 D2 stop rule):
- `kaylousberg.itch.io/kaykit-adventurers`: **Adventurers 2.0**. Its free tier adds a **Ranger**.
- `kaylousberg.itch.io/kaykit-character-animations`: **Character Animations 1.1**. 150+ free clips on
  `Rig_Medium`, advertised as compatible with all KayKit characters.

The GitHub archives exclude the FBX sources through `export-ignore`. That doesn't matter, because
every character, weapon and dungeon piece ships as `.glb` or `.gltf`.

## 1. License — **resolved: CC0**

All three staged packs carry a `LICENSE.txt` stating CC0 ("free to use in personal, educational and
commercial projects", crediting optional). That is on the ADR-0018 D2 allow-list. The itch.io pages
for Adventurers 2.0 and Character Animations 1.1 also state CC0. Their archive text still has to be
checked once the owner downloads them.

## 2. Contents and class mapping — **resolved, with gaps**

Adventurers 1.0 contains **five** characters, not the four its README says: Barbarian, Knight,
Mage, Rogue, Rogue_Hooded.

| Class | Kit character (1.0) | Kit character (2.0, owner download) | Notes |
|-------|---------------------|-------------------------------------|-------|
| paladin | Knight | Knight | 4 shields, 1H/2H sword, helmet |
| barbarian | Barbarian | Barbarian | 1H/2H axe, round shield, hat |
| sorcerer | Mage | Mage | wand, staff, spellbook, hat |
| rogue | Rogue | Rogue | knives, crossbows, throwable |
| ranger | Rogue_Hooded | **Ranger** | 1.0 has no bow; 2.0 Ranger is the proper fit |

**Current `item_visuals` assets mapped to kit accessories.** The accessories are nodes that already
exist under the `handslot.r` / `handslot.l` / `head` bones, or standalone `.gltf` files in the pack.

| Slot | Current asset(s) | Kit equivalent | Gap |
|------|------------------|----------------|-----|
| main_hand | `weapon_rusty_sword_v0`, `weapon_long_sword_v0` | `1H_Sword` / `sword_1handed`, `2H_Sword` / `sword_2handed` | — |
| main_hand | `weapon_rapier_v0` | `sword_1handed` or `dagger` | approximate (no rapier) |
| main_hand | `weapon_starter_axe_v0` | `1H_Axe` / `axe_1handed`, `2H_Axe` | — |
| main_hand | `weapon_starter_staff_v0` | `2H_Staff` / `staff`, `1H_Wand` / `wand` | — |
| main_hand | `weapon_training_bow_v0` | **none in 1.0** (only crossbows, `arrow`, `quiver`) | **bow gap**: 2.0 Ranger, or a crossbow stand-in |
| off_hand | `equipment_shield_kite_v0`, fallback | `shield_badge`, `shield_square`, `shield_round`, `shield_spikes` (+ `_color` variants) | — |
| head | fallback helmet | per-class headgear: `Knight_Helmet`, `Barbarian_Hat`, `Mage_Hat`, `Rogue_Head_Hooded` | the head visual varies by class, not by item |
| chest, gloves, boots, belt | fallbacks | **tints** (D5); see §4 | belt has no own region |
| ring, amulet | fallbacks | none (UI only, D5) | by design |

Monsters: Skeletons 1.0 provides Minion, Warrior, Mage and Rogue, plus a standalone axe, blade,
staff, crossbow, shields and arrows. That covers `dungeon_undead`. Beasts (wolf, fox, bat) and the
demon or purple monster have **no KayKit 1.0 equivalent**, so the P4 decision stands.

## 3. Shared rig — **resolved for 1.0; 2.0 still unverified**

`report.md` shows **one rig group**: all 5 Adventurers and all 4 Skeletons share an identical
**41-joint** set. The deform chain is:
`root → hips → spine → chest → {upperarm, lowerarm, wrist, hand, handslot}.{l,r}`, plus `head` and
`{upperleg, lowerleg, foot, toes}.{l,r}`. IK and control helper joints come on top.

**The clips are embedded in each character file: 76 per Adventurer and 95 per Skeleton.** The
core animation set therefore does not depend on the Character Animations pack. The probe confirms
`AnimationPlayer` binds to the 1.0 rig without retargeting.

**Socket and camera bones**, replacing the canonical 17-bone names in `gear_sockets.v0.json`
during P3:

| Socket / anchor | Kit bone |
|-----------------|----------|
| `right_hand_socket` | `handslot.r` |
| `off_hand_socket` | `handslot.l` |
| `head_socket` | `head` |
| `chest_socket` | `chest` |
| `chest_view` first-person camera anchor | `head` |

**Logical animation state to kit clip** (same names on Adventurers and Skeletons):

| State | Kit clip | Notes |
|-------|----------|-------|
| idle | `Idle` | `2H_Melee_Idle` when a 2H weapon is equipped |
| walk | `Walking_A` | `Running_A` is also available; the speed threshold goes in the P3 catalog |
| attack | `1H_Melee_Attack_Chop` | also `_Slice_Diagonal`, `_Slice_Horizontal`, `_Stab` |
| attack_off_hand | `Dualwield_Melee_Attack_Stab` | or `1H_Melee_Attack_Stab` |
| attack_2h | `2H_Melee_Attack_Chop` | also `_Slice`, `_Spin`, `_Stab` |
| attack_ranged | `2H_Ranged_Shoot` / `1H_Ranged_Shoot` | crossbow motion. **Gap: no bow draw** |
| attack_staff | `Spellcast_Shoot` | also `Spellcast_Raise`, `Spellcast_Long` |
| hit | `Hit_A` | `Hit_B` |
| death | `Death_A` | then hold `Death_A_Pose` |

Useful extras for later slices: `Block*`, `Dodge_*`, `Jump_*`, `PickUp`, `Interact`, `Use_Item`,
`Cheer`. Skeletons also have `Spawn_Ground_Skeletons`, `Skeletons_Awaken_*` and
`Death_C_Skeletons_Resurrect`.

**Open risk carried into P3:** whether Adventurers 2.0 and Character Animations 1.1 (`Rig_Medium`)
use the same 41-joint names as 1.0 is **unverified**. `make inspect-kit` on the owner-downloaded
archives answers it in one run. If they differ, the choice is to stay on 1.0 (bow gap) or retarget.

## 4. Mesh split and tint mechanism — **resolved: mechanism 1 (per-part material override)**

Every character is split into six named body meshes:
`<Class>_Body`, `_ArmLeft`, `_ArmRight`, `_LegLeft`, `_LegRight`, `_Head` (for example
`Knight_Body`; Rogue_Hooded uses `Rogue_Head_Hooded`). Skeletons follow the same pattern
(`Skeleton_Warrior_Body`, …). Accessories are separate mesh nodes parented to bones.

The probe applied a duplicated `StandardMaterial3D` with a tinted `albedo_color` as
`material_override` on `Knight_Body` and both legs, and it rendered correctly. No shader was
needed. Mechanism 2 (region-mask shader) is **not needed**.

**What each tint can actually reach** (the owner accepted this granularity):

| Armor slot | Tint target | Honest limit |
|------------|-------------|--------------|
| chest | `<Class>_Body` | also covers the belt area and trim, since it is one mesh |
| gloves | `<Class>_ArmLeft` + `_ArmRight` | **whole arms**; hands cannot be isolated |
| boots | `<Class>_LegLeft` + `_LegRight` | **whole legs**; feet cannot be isolated |
| belt | none of its own; shares `<Class>_Body` | **no separate belt visual**. P3 picks: belt tint only when no chest item, or no belt visual |

Albedo multiply also darkens the atlas's own colors. P3 may prefer a lerp toward the tint over a
pure multiply; that is a presentation-catalog value, not a new mechanism.

## 5. Tile size vs server walls — **resolved: client-only P2 (no server snap slice)**

Kit modules measured by `inspect-kit`, in kit units:

| Piece | Size |
|-------|------|
| `wall` | 4 × 4 × 1 (length × height × depth) |
| `wall_half` | 2 |
| `wall_endcap` | ≈1.07 |
| `wall_corner` | 2.5² |
| `wall_Tsplit` / `wall_crossing` | available |
| `floor_tile_small` | 2 × 2 |
| `floor_tile_large` | 4 × 4 |
| `pillar` | 1.5 × 4 |
| `column` | 0.7 |

Wall-grid audit: 20 seeds × levels −1..−10 (every floor profile, every biome, two boss floors),
22,312 wall edges.

| Grid | Edges aligned | `room_wall` | `perimeter` |
|------|---------------|-------------|-------------|
| 0.5 | **100%** | 100% | 100% |
| 1.0 | 91.6% | 93.1% | 100% |
| 2.0 | 45.7% | 46.3% | 50% |
| 4.0 | 23.2% | 22.7% | 30% |

Wall thickness is overwhelmingly 1.0 (5,086 of 5,578 walls).

**Outcome:** render the environment at **XZ scale 0.5**:
- a kit half wall (2) becomes **1.0** game units, the grid that 91.6% of edges already sit on
- a kit end cap (~1) becomes ~0.5, covering the remaining half-unit offsets. Every edge is on the
  0.5 grid
- a small floor tile becomes 1 × 1
- wall depth 1 × 0.5 = 0.5 is stretched ×2 to the server's 1.0 thickness

Water and holes (54–62% on the 1.0 grid) are floor features rendered as surfaces, not wall pieces,
so their alignment does not constrain the tiler. **No server snap rule is required.** One can be
added later only if P2 captures show seams.

**Vertical scale:** `ceiling_height` is 4.0. Current heroes are ~1.9 tall, and the kit Knight is
2.47 (helmet included). The proposal for P2 and P3, to be confirmed with captures:
- characters at about **×0.77** (keeps today's hero height and gameplay readability)
- environment Y at about **×0.75** (walls about 3.0 tall)

## 6. Budgets — **resolved (measured)**

| Asset class | Measured (kit) | Final ADR-0018 D8 budget |
|-------------|----------------|--------------------------|
| Hero / humanoid monster | 4.6k–7.0k tris per file, **including every hidden accessory mesh**; one 1024² atlas | ≤ 10k tris, ≤ 1024² |
| Large monster / boss | n/a in kit | ≤ 20k tris, ≤ 1024² (unchanged) |
| Weapon / shield / helmet | 84–830 tris | ≤ 3k tris (unchanged) |
| Dungeon piece / prop | most < 1.2k; max **3,499** (`wall_cracked` 2,010, `stairs_wood_decorated` 2,537) | **≤ 4k tris** (raised from 2k) |
| Textures | every kit file uses one **1024²** atlas (Adventurers, Skeletons: per pack; Dungeon: shared) | ≤ 1024² per asset |

## Probe (D2)

`.artifacts/kaykit/probe/`, a throwaway Forward+ project, never committed, loaded `Knight.glb` and
`Skeleton_Warrior.glb` at runtime through `GLTFDocument` (no import step). Results:
- `AnimationPlayer` found. Knight has 76 clips (posed at `1H_Melee_Attack_Chop` 0.45 s); Skeleton
  has 95 (`Walking_A` 0.4 s).
- The handslot accessories toggled to exactly `1H_Sword` + `Round_Shield`.
- `Knight_Body` and both legs were tinted.
- `handslot.r` / `handslot.l` / `head` / `chest` / `hips` resolved on `Skeleton3D`.

Capture: [`kaykit-rig-probe.png`](../as-built/assets/v469/kaykit-rig-probe.png).

## Other findings from this slice

1. **Dungeon generation fails on ~2% of floors.** 4 of 200 seed/level pairs fail with
   `could not place room-corridor layout after 96 attempts`: (`audit-09`, −1), (`audit-13`, −3),
   (`audit-17`, −7), (`audit-20`, −3). `Sim.ensureDungeonLevel` returns the error, so a session
   with that seed can never enter that floor. This is out of scope here; a follow-up task was filed.
2. **Two screenshot focuses were silently broken since `ba083f77`.** `town` and `stairs` called
   `main.gd` functions that had moved to `TownNodeFactory`. They saved blank frames and passed as
   "ok". Both are fixed, and `render_focus.py` now fails a capture on any `SCRIPT ERROR` in the
   Godot log.
3. **The dungeon-room capture shows the water pool as pale streaks on a wall face** (see the
   `dungeon-room-*` baselines). This is a current presentation bug; P2 replaces that renderer.
4. **The `monsters` focus renders every GLB monster untextured white** (see the baseline). This is
   the current look; P4 replaces those models.
5. **Environment:** macOS `/usr/bin/make` (and `strings`) refuse to run until the Xcode license is
   accepted. Local Godot is 4.7.2, while PROGRESS.md pins 4.6.3.

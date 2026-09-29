# v476 — CC0 beasts: Quaternius wolf and bat (ADR-0018 P4b)

- **Status:** Implemented (v476)
- **Date:** 2026-09-29
- **Codename:** `cc0-beasts`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D1/D2/D8, phase P4b
- **Baseline:** v475 (`a76953f7`)

## Purpose

1. **Replace the last three unconfirmed-license assets** (ADR-0018 D2: "after P4, none"):

   | Monster def | Scene key (unchanged) | Old model | New model |
   |-------------|-----------------------|-----------|-----------|
   | `dungeon_wolf` (Cave Wolf, pounce) | `monster_quadruped` | evil fox, 39.2k tris, 2048² texture, budget exemption | Quaternius **Wolf** |
   | `companion_black_wolf` + skill wolf companions | `monster_wolf` | user-provided wolf | Quaternius **Wolf** (same GLB) |
   | `dungeon_bat` (Cave Bat, dive) | `monster_tiny_flyer` | Sketchfab low-poly bat | Quaternius **Bat** |

   Source: Quaternius on Poly Pizza, **CC0 1.0**. Wolf and Fox come from the *Animated Animal Pack*
   (`poly.pizza/m/P1gU3Qkr9r`, 1,962 tris, 51 joints). The Bat is `poly.pizza/m/hNO9XvjlKa` (1,046
   tris, 23 joints). Both have no textures (material colours only) and embedded clips.
   Owner decision (2026-09-29): enemy and companion share the wolf GLB and are told apart by the
   existing companion tint (`101014`/`202028`) and `visual_scale` data in `skills.v0.json`. The Fox is
   staged under `.artifacts/quaternius/` but not vendored.
2. **Scene keys stay.** Keeping `monster_quadruped` / `monster_wolf` / `monster_tiny_flyer` means no
   change to the server boss whitelist, `boss_templates`, `skills.v0.json`, goldens or bot scenarios.
   Each scene becomes a `KitMonsterVisual` root (the v474 recipe) with a clip profile in
   `kit_monster_presentation.v0.json`:

   | Logical clip | Wolf | Bat |
   |--------------|------|-----|
   | idle | `Idle` | `BatArmature\|Bat_Flying` |
   | walk | `Walk` | `BatArmature\|Bat_Flying` |
   | attack | `Attack` | `BatArmature\|Bat_Attack` |
   | hit | `Idle_HitReact_Left` | `BatArmature\|Bat_Hit` |
   | death | `Death` | `BatArmature\|Bat_Death` |
   | pounce / dive | `Gallop_Jump` (pounce) | `BatArmature\|Bat_Attack2` (dive) |
3. **Purge.** Remove the fox, old wolf and old bat everywhere:
   - runtime GLBs, texture sidecars and source copies
   - the three `.tres` libraries
   - manifest entries and the fox budget exemption
   - `tools/assets/rig_quadruped_monster_glbs.py`, its test and its `gen-assets` line
   - `_wolf_clips` in `client/tools/build_animations.gd`
   - the dead `monster_tiny_flyer_glb()` in `gen_glb.py`
   - the stale READMEs
4. **License gate.** `validate_assets.py` rejects any manifest license outside an allow-list (`CC0-1.0`,
   `CC-BY-3.0`, `CC-BY-4.0`), so an unconfirmed asset can't come back.
5. **Stale whitelist.** `server/internal/game/rules.go` boss `model_pool` still lists the models
   purged in v474 (`monster_dark_purple`, `monster_crocodile_archer`). Remove them. No live data uses
   them, so gameplay is unchanged.

## Non-goals

- No new monster defs or scene keys. Kit skeleton scenes are not added to the boss `model_pool`.
- No gameplay, rules, protocol or golden changes.
- No per-biome beast variants, and no second beast species.

## Acceptance criteria

1. **Assets.** `quaternius_wolf_v0` and `quaternius_bat_v0` are `type: monster` manifest entries with
   CC0 provenance (`origin`, `source_url`, sha256). `required_nodes` are real skin joints. Both are
   within the D8 budgets with no exemption. The CC0 legal text is committed next to them.
2. **Scenes.** Each of the three scenes has a model, an `AnimationPlayer` holding every logical clip in
   its profile (including `pounce` / `dive`), a nose facing the same way as the kit skeletons, and a
   size comparable to the old model.
3. **Mapping.** `monster_visuals.v0.json` maps `dungeon_wolf` and `companion_black_wolf` to
   `quaternius_wolf_v0`, and `dungeon_bat` to `quaternius_bat_v0`. Scale and height offset are kept or
   retuned.
4. **Purge complete.** Nothing references `evil_fox`, `purple_fantasy`, `wolf.glb`, `tiny_flyer.glb`,
   `monster_quadruped_predator_v0`, `monster_wolf_v0` or `monster_tiny_flyer_v0`, except history docs.
   `make validate-assets` shows no orphans and no unconfirmed license.
5. **Visuals.** The `monsters` / `companions` showme focus is shown before/after next to a KayKit
   skeleton (ADR-0018 coherence rule).
6. **Gates.** `make ci` passes. Client tests covering beasts (`test_animation`, `test_kit_monsters`,
   `test_coop_client`, `test_model_viewer`) assert on the new rigs.

## Asset decision

- **Adopt** Quaternius Wolf + Bat (CC0). These are low-poly and animated with the needed clips, and
  come in at ~5% of the old fox's triangles.
- **Style risk (accepted pending the capture).** Quaternius shades flat with realistic wolf
  proportions, while KayKit is smooth-shaded chibi. The bat reads close to KayKit. The wolf is
  acceptable at isometric distance. The owner can reject it on the as-built capture.
- **Reject** keeping the unconfirmed assets, and reject the Fox for the companion (owner chose one
  GLB).

## Risks

- **Scale.** The GLB armatures carry a 100× node scale. The scene transform must be tuned against the
  kit skeleton, and the pick collider and `height_offset` data must be checked.
- **Duplicate clips.** Each clip is exported twice (bare and `AnimalArmature|`-prefixed). That's
  harmless (~1 MB per GLB), and the profile uses the bare names.

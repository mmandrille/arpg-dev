# v511 — Monster variant looks

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)
- **Date:** 2026-10-01
- **Codename:** `monster-variant-looks`
- **Area:** Client presentation / graphics (presentation-only)
- **Base commit:** `425b9ae4`
- **Depends on:** v501 monster death dissolve (rim flash owns emission during death) and v507 class silhouettes; both are already on the base. No unintegrated prerequisite.
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D1/D2/D7 (art direction, sourcing, budgets); [ADR-0007](../adr/0007-animation-state-model.md) (client-only presentation); [ADR-0001](../adr/0001-technology-stack.md) D2 (server authority).

## Purpose

Monsters look identical. At the base there are five staged families: three KayKit skeletons (`dungeon_mob`, `dungeon_archer`, `dungeon_undead`), the Quaternius wolf (`dungeon_wolf`) and bat (`dungeon_bat`). A champion today differs from a common only by the server-sent `visual_tint` (a flat `albedo_color` multiply, which dulls a textured atlas) and `visual_scale` (champion 1.25, unique 1.5; rare 1.0, so a rare looks like a pink common). Depth changes nothing.

Make variants readable at a glance, driven by data in `shared/assets/kit_monster_presentation.v0.json` (+ its schema), by composing three axes:

1. **Depth** — the already-resolved biome palette id (`shallow_cave`, `sundered_halls`, `deep_vault`) shifts a family's tint so the same skeleton reads differently by depth.
2. **Rarity** — `common`, `champion`, `rare`, `unique` each get a distinct tint strength, scale multiplier, eye glow, and (champion and above) a ground aura ring.
3. **Family** — per-family overrides (eye bone/offset, aura radius), because rig and size differ (skeleton vs wolf vs bat).

Server authority is unchanged: rarity, depth, `visual_scale`, and `visual_tint` remain server/wire inputs; this slice only changes how the client renders them.

## Non-goals

- No server, protocol, rules, golden, or replay change. No new wire field (rarity, `visual_scale`, `visual_tint`, `monster_pack_leader` already ship on the entity).
- No new assets, packs, or models (ADR-0018 D2: already-staged KayKit/Quaternius only). No new GLB; eye glow and aura are procedural primitives.
- No new monster families, clips, or animation work (v513), and no boss-specific presence (v512). Bosses keep their existing explicit `visual_tint`/`visual_model` path untouched.
- No change to elite-command aura lights (`aura_soft_lights.gd`) or pack-leader preview; the champion ring is a different, cheaper, always-on cue and must not replace them.
- No per-monster `OmniLight3D`/`SpotLight3D`, no shader/plugin/post-process addition, no new `GPUParticles` per monster.
- No retuning of damage, HP, loot, or `monster_rarities` gameplay columns.

## Observable acceptance criteria

- [ ] `kit_monster_presentation.v0.json` gains a schema-validated `variants` section (rarity looks, depth looks, per-family overrides). Schema is `additionalProperties: false`; `validate_shared.py` and `validate_assets.py` check every rarity id against `dungeon_generation.v0.json` `monster_rarities`, every depth key against the dungeon `biome_palettes` ids, and every family key against `monster_visuals` families with a kit/Quaternius scene. Unknown or missing entries fall back to the common/neutral look (documented safe base), never an error at runtime.
- [ ] A pure resolver `(visual_key, rarity, palette_id) -> VariantLook` exists with no scene-tree dependency. Semantic tests (Test Locking Policy), derived from the loaded catalog rather than hardcoded values: for each of the five families, every rarity pair resolves to a distinct look (tint or scale or eye or aura differs); resolution is deterministic; an unknown rarity/palette resolves to the neutral look; depth looks differ across the three palettes for at least the skeleton family.
- [ ] Applying a look to a real family scene (all five) sets tint via the material **detail layer** (`ModelDetailTint`), leaves `albedo_color` to status/hit-flash owners (`ModelReactionController.set_base_tint`), and preserves the albedo texture. Burn/poison/bleed/ice status tints and the hit flash still compose and restore (existing behaviors tested).
- [ ] Eye glow: eligible families (data-listed) show emissive eye markers on a catalog-named bone; common may have none, champion/rare/unique have distinct colors/energies. Markers are shared-mesh unshaded/emissive primitives, no lights. Death (v501 rim flash) and dissolve are unaffected: markers hide/fade with the corpse and do not leave floating glow after death.
- [ ] Champion aura: champion/rare/unique show a ground ring whose color/radius/alpha come from data and scale with the family; common has none. Elite-command aura lights still render independently.
- [ ] Scale: effective visual scale = server `visual_scale` × catalog `scale_multiplier` (rare gets a data-driven bump from 1.0; unique/champion default to 1.0 to avoid double scaling). Health bar, hit highlight, and click targeting still line up with the scaled model (verified in captures and the existing client tests).
- [ ] Back-compat: monsters with no catalog entry, bosses (`visual_model` override), companions/mercenaries, training dummies, and lab monsters render exactly as at the base (asserted by test: same material/scale/children as before the slice).
- [ ] Real-renderer showme captures: one row per family × rarity (5 × 4) at one depth, plus the skeleton warrior at all three depth palettes, color and grayscale; plus an in-game player-camera capture of a mixed-rarity pack. Review confirms rarity ordering is distinguishable in grayscale (not hue alone, consistent with v508 color-safe cues) and silhouettes stay readable against the dungeon floor.
- [ ] Performance gate (budgets set before tuning, inheriting v498/v505: added frame/process p95 ≤ `max(0.5 ms, 5%)`, draw calls ≤ `max(10, 5%)`, first-spawn cost ≤ `max(20 ms, 10%)`): a matched run on `sorcerer_multigroup_perf_probe` / `benchmark_mixed_arena` (v495/v497 fixture) A/B against the base. Material/mesh count stays bounded: variant materials come from a capped cache keyed by `(source material, look)` so 5 families × 4 rarities × 3 palettes cannot exceed the cap and thrash (the v495 cache cap of 64 is raised or split with a recorded reason). No per-frame allocation in the variant path; applied once on node creation/rarity change.
- [ ] Validators and focused tests pass; `make maintainability` passes with no grandfathered file growth (new logic lives in new files).

## Likely surfaces (adopt / borrow / reject)

Existing in-repo code inspected: `model_tint.gd` (albedo multiply with a 64-entry cache), `model_detail_tint.gd` (texture-preserving lerp via the detail layer; owns no `albedo_color`), `monster_family_accent.gd` (ground torus, the pattern for the aura ring), `kit_monster_visual.gd` (applies clips + bone attachments from this catalog), `kit_monster_presentation_loader.gd`, `model_reaction_controller.gd` (owns `albedo_color` base tint, hit flash, highlight and death-rim emission), `main.gd` `_make_entity_node` / `_entity_base_tint` / `_apply_model_tint`, `monster_visuals_loader.gd`, v505 `dungeon_depth_mood_loader.gd` + `DungeonDepthLighting.palette_id_for_level` (palette id resolution to reuse).

- **Adopt:** `ModelDetailTint` for variant tint (composes with `albedo_color`, keeps atlas shading — the explicit reason a multiply can only darken); the `MonsterFamilyAccent` ring pattern for the champion aura; the BoneAttachment3D mounting pattern from `kit_monster_visual.gd` for eye markers; the v505 palette-id resolution (no duplicated depth cutoffs in client code); the loader pattern (`class_name`, static `ensure_loaded()`).
- **Borrow:** the one-shared-material-per-look cache idea from `ModelTint`/v495.
- **Reject:** per-monster lights (v497 frame pacing), a new shader/plugin, new art packs (ADR-0018 D2), and a second tint path that overwrites `albedo_color` (breaks status/hit flash composition).
- **Data:** extend `shared/assets/kit_monster_presentation.v0.json` with a top-level `variants` object and `kit_monster_presentation.v0.schema.json` accordingly. Gameplay tuning stays in `dungeon_generation.v0.json`; the slice only reads `rarity` ids from it in validators.
- **Client (new files):** `monster_variant_resolver.gd` (pure), `monster_variant_look.gd` (apply/remove on a node: detail tint, eye markers, aura ring), plus loader accessors in `kit_monster_presentation_loader.gd`. Narrow hooks in `main.gd` (`_make_entity_node` after the model is built and when a record's rarity/palette changes) and `kit_monster_visual.gd` only if the apply hook lives there. `model_reaction_controller.gd` should not need edits; if it must (e.g. detail-tint capture on private material), keep it to a few lines and note it for v513.
- **Tools/tests:** `tools/validate_shared.py` + `tools/assets/validate_assets.py` cross-checks (with a python test); a new `client/tests/test_monster_variant_looks.gd` registered in `scripts/client_smoke.sh`; extend `test_kit_monsters.gd` only where it already owns the contract; showme: extend the existing `monsters` focus (`client/scripts/showme/visual_capture.gd` `_setup_monsters`) with a rarity/depth matrix and a regen-screenshots suite entry; a client bot scenario only if a live pack capture needs a placement not available from lab worlds.
- **Docs:** `docs/CODEMAP.md` Assets and Hero-light-and-visuals rows; spec/plan/as-built; lifecycle rows are the coordinator's.

## Focused verification

```bash
make validate-shared validate-assets
python3 -m pytest tools/assets/test_validate_assets.py tools/test_validate_shared.py -q
make client-unit            # includes test_monster_variant_looks, test_kit_monsters, test_look_and_feel_polish
make maintainability && git diff --check
make regen-screenshots SUITE="monsters"     # suite name finalized in the plan
```

Bot/real-camera proof: one client scenario (headless + one visible renderer run) that places a mixed-rarity pack of two or more families in the actual player camera and asserts via the existing debug state that each monster node carries the expected variant markers; real-renderer (Metal/Forward+) captures per the acceptance list. Headless `gl_compatibility` does not prove the final look.

## Dependencies and integration risks

- No gameplay prerequisite. Siblings v512 (boss presence) and v513 (monster animation polish) touch the same presentation files (see the plan's conflict table); v512 should consume this catalog's `unique` look rather than invent a parallel one.
- Risk: the server already sends `visual_tint` per rarity, which `_entity_base_tint` applies as an albedo multiply. If the variant catalog also tints, the result double-applies. The plan resolves this by making catalogued families use a neutral base albedo and the variant detail tint as the sole rarity color (see decision D2); non-catalogued monsters and bosses keep today's path.
- Risk: eye bones differ between the KayKit skeleton rig and the Quaternius rigs; the plan starts with an inventory task and falls back to a head-bone attach, or omits the eye glow for a family, rather than guessing offsets.
- Risk: first-spawn hitch (v495) — variant application must not add per-monster material duplication; measured by the A/B gate.
- Risk: `main.gd` and `validate_shared.py` are grandfathered over-limit files; edits are minimal call-throughs.

## Open questions affecting planning

1. **Scope of server tint:** may the client ignore the server `visual_tint` for catalogued rarity monsters (neutral base + catalog look), keeping `visual_tint` only for bosses/companions? Recommended yes; the server value equals the rarity table color and would still be used for UI/legacy fallbacks. (Presentation-only; no server edit.)
2. **Depth axis granularity:** three palette ids (reusing v505) versus per-depth bands. Recommended palette ids; confirm.
3. **Wolf/bat eye glow:** include if the rig exposes a head bone cheaply; otherwise skeleton-only eyes in this slice. Plan decides after the inventory task; no user input needed unless art direction objects.

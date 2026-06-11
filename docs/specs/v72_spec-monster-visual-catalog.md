# v72 Spec — Monster Visual Catalog

Status: Complete
Date: 2026-06-11
Codename: `monster-visual-catalog`

## Purpose

Add a clean data-driven presentation path for monster models, then prove it with two new enemy
visual families: a grounded quadruped predator and a tiny hovering flyer. Monster mechanics remain
server-authored through existing monster definitions; the client resolves scenes, scale, height
offset, and animation profile from shared visual metadata instead of adding one-off GDScript
branches per monster.

The slice also removes the current boss visual special case: generated bosses may use any supported
monster visual, including the existing dummy, the new quadruped, or the new tiny flyer, while
remaining deterministic from the server/session data.

## Non-goals

- No true flying gameplay, vertical collision, wall bypass, or pathing changes.
- No new monster AI behavior such as pounce, swarm, dive attack, dodge, or ranged bat combat.
- No protocol schema change unless implementation discovers that existing entity metadata is
  insufficient; the preferred path is shared visual data plus existing `monster_def_id`,
  `visual_model`, `visual_scale`, and boss metadata.
- No production-quality art pass. Models are deterministic low-poly placeholders that establish
  silhouette, scale, animation hooks, and the data contract.
- No loot or economy rebalance beyond reusing existing dungeon mob rewards.

## Acceptance Criteria

- `shared/assets/monster_visuals.v0.json` and schema define monster presentation by data:
  `monster_def_id` or visual key to asset id, scene path/key, scale, height offset, and animation
  profile.
- Every dungeon/boss monster definition introduced or used by normal generation has a valid visual
  mapping, and validation fails on missing or unknown asset ids.
- New monster definitions exist for `dungeon_wolf` and `dungeon_bat`; their combat behavior stays
  mechanically equivalent to existing chase dungeon mobs for this slice.
- Dungeon generation can spawn `dungeon_mob`, `dungeon_archer`, `dungeon_wolf`, and `dungeon_bat`
  through data weights, with deterministic minimum coverage where tests require it.
- Boss generation can deterministically select or inherit any of the three supported model families:
  original dummy/biped, quadruped, or tiny flyer. The client renders the selected visual without a
  hardcoded boss-only model branch.
- The Godot client uses a reusable monster visual resolver/loader rather than scattered
  `monster_def_id` conditionals for model choice. Archer bow marker compatibility may stay as a
  narrow overlay until ranged-monster presentation is generalized.
- New scenes for the quadruped and tiny flyer load in headless client tests and expose required
  animation clips: `idle`, `walk`, `hit`, and `death`.
- The tiny flyer visually hovers above the ground and flaps/floats in idle/walk presentation, while
  authoritative server position remains on the existing 2D ground plane.
- The showme workflow supports a focused monster lineup capture. The implementation must run the
  capture and inspect it before broad wiring; if scale or silhouette is questionable, stop for user
  feedback before finalizing the models.
- Bot/client proof covers a lab world where the wolf and bat are visible as distinct monster
  definitions, and existing combat/boss scenarios continue to pass.

## Scope and Likely Files

Shared contracts and rules:
- `shared/assets/monster_visuals.v0.schema.json`
- `shared/assets/monster_visuals.v0.json`
- `assets/manifests/assets.v0.json`
- `assets/manifests/assets.v0.schema.json` if validation metadata needs tightening
- `shared/rules/monsters.v0.json`
- `shared/rules/dungeon_generation.v0.json`
- `shared/rules/worlds.v0.json`
- `shared/rules/boss_templates.v0.json` and schema if boss visual pools move into boss data
- `shared/golden/*` for deterministic dungeon/boss visual generation as needed

Asset pipeline and tools:
- `tools/assets/gen_glb.py`
- `tools/assets/validate_assets.py`
- `tools/assets/test_validate_assets.py`
- `client/assets/monsters/**`
- `assets/monsters/**/README.md`

Godot client:
- `client/scripts/main.gd`
- New `client/scripts/monster_visuals_loader.gd` or equivalent
- New/updated monster scenes under `client/scenes/`
- `client/tests/test_animation.gd`
- `client/tests/test_item_visuals.gd` or new monster-visual test

Bot and showme:
- `tools/bot/scenarios/*.json`
- `tools/bot/scenarios/client/*.json` if client visual debug proof is used
- `skills/showme/scripts/render_focus.py`
- `skills/showme/scripts/visual_capture.gd`

Docs:
- `docs/plans/v72_2026-06-11-monster-visual-catalog.md`
- `docs/as-built/v72_monster-visual-catalog.md`
- `PROGRESS.md`

## Test and Bot Proof

- `make validate-shared` validates new shared monster visual metadata and rule additions.
- `make validate-assets` validates monster asset ids, runtime GLBs, and provenance.
- Targeted Go tests prove deterministic dungeon and boss model selection if generation changes.
- Godot headless tests load all monster scenes and verify required clips.
- A focused showme capture renders a monster lineup containing dummy, wolf, bat, and boss-scale
  variants before the visual assets are treated as final for this slice.
- A protocol/client bot scenario proves the new monster definitions are present and still playable
  under existing authoritative mechanics.
- Final gate: `make ci`.

## Open Questions and Risks

- **Visual approval risk:** deterministic generated models may need silhouette or scale tuning.
  The plan must include a showme feedback loop before finalizing the model files.
- **Boss data placement:** the implementation should prefer a data-driven boss visual pool in
  shared rules if the current boss template visual block cannot express all three families cleanly.
- **Scope control:** archer bow presentation is currently a special overlay. This slice may leave it
  as a compatibility exception if fully general ranged-weapon overlays would exceed v72.

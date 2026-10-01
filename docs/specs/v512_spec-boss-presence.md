# v512 Spec — Boss Presence

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)
- **Date:** 2026-10-01
- **Codename:** `boss-presence`
- **Batch base:** `425b9ae4` (detached v512 worktree `/Users/mmandrille/git/arpg-dev-batch/v512-boss-presence`)
- **Depends on:** v509 boss lane telegraphs (on main), v474 KayKit monster scenes, v287 two-template boss set.
  No unfinished sibling slice is a prerequisite; recheck the integrated baseline (v511, v513) before execution.
- **ADRs:** ADR-0009 (boss floors; D-series telegraphs, health bar), ADR-0018 (kit visuals, tint rules,
  render quality tiers, screenshot gate), ADR-0001 D2 (server authority), ADR-0007 (animation is client-only).

## Purpose

Bosses should read as bosses, not as a big skeleton. Today:

- Cave Warden is sent by the server with `visual_model` drawn at random from
  `monster_dummy | monster_quadruped | monster_tiny_flyer` (`boss_templates.v0.json` `model_pool`,
  chosen in `dungeon_population.go`), so the first boss can be a scaled wolf, a bat or a dummy.
- Crypt Matron is the legacy `monster_skeleton` GLB at scale 2.2 (Cave Warden: 2.0), nearly identical size.
- The only boss-specific world dressing is one shared red `BossArenaPresence` torus (`DEFAULT_COLOR`,
  fixed radius 2.4) and the screen-top boss health bar. Nothing announces the encounter.

v512 gives each boss template a presentation identity owned by a schema-backed client catalog:
a fixed kit model, a distinct scale, a headgear and weapon attachment, a themed ground aura, and a
one-shot intro name banner when the player arrives at a live boss. Pure client presentation; the
server keeps owning stats, patterns, positions, telegraph geometry, and hit resolution.

## Product decisions recorded (defaults; see Open questions)

1. **Catalog home:** new `shared/assets/boss_presentation.v0.json` (+ `.schema.json`), keyed by
   `boss_template_id`. It owns presentation only. `shared/rules/boss_templates.v0.json` keeps
   gameplay fields; its `visual` block stays as the wire fallback (server events, goldens, protocol
   scenarios 84/95 keep passing unchanged).
2. **No server or protocol change.** `boss_template_id`, `is_boss`, `hp`, `max_hp`, `visual_scale`,
   `visual_model` are already on the entity wire. The client overrides model/scale by
   `boss_template_id`. Justified under ADR-0001 D2: nothing here is an outcome.
3. **Assets:** already-staged only. Skeleton Mage rig and Skeleton_Staff / Skeleton_Shield_Large_*
   come from the staged Skeletons 1.0 archive (`.artifacts/kaykit`, same pack and CC0 license as the
   v474 monsters, packed with `tools/assets/gltf_to_glb.py`). No new pack, no download.
4. **Banner trigger:** client-side, once per boss entity id, the first time a live full-health boss
   appears in the entity map after a level change/snapshot (boss-floor entry). Reconnect into a fight
   in progress (hp < max_hp or active `boss_phase`) does not replay it.

## Non-goals

- No new boss template, pattern, stat, loot, hit shape, or arena layout; no balance change.
- No server, protocol, golden-fixture, or `boss_templates.v0.json` gameplay-field edit
  (removing the stale `model_pool` is a follow-up; it would churn seeded visual RNG and goldens).
- No new art packs, no imported VFX plugin, no audio, no boss portraits/animated boss art.
- No change to health bar layout/behavior (v517 owns HUD/health-bar work); v512 only supplies the
  banner and reads the same record fields.
- No per-boss new animation clips (v513 owns monster animation polish); reuse the existing
  `kaykit_skeleton_melee` / ranged clip profiles.
- No co-op banner synchronization beyond each client showing its own once-per-entity banner.

## Observable acceptance criteria

1. **Catalog + schema.** `boss_presentation.v0.json` validates against its schema. Each boss in
   `boss_templates.v0.json` has exactly one entry with: `display_name` and `epithet` (banner),
   `visual_key` (an entry in `kit_monster_presentation.v0.json`), `scale`, optional `headgear`,
   `weapon`/`attachments`, `aura`, and `banner` timing/colors. `validate_shared.py` fails on a missing
   entry, unknown manifest asset id, unknown bone, non-positive scale, banner duration outside
   [1.0, 6.0] s, bad hex color, or two bosses whose scales differ by less than 15%.
2. **Distinct model and scale.** Cave Warden renders as a kit skeleton warrior-class model and
   Crypt Matron as a kit skeleton mage-class model, never a wolf/bat/dummy/player model, regardless
   of the server `visual_model`. Their effective scales differ visibly (ratio >= 1.15) and both
   exceed any normal-rarity monster. Arena presence, telegraph markers, hit-reaction roots and the
   health-bar anchor use the effective scale (no double scaling).
3. **Headgear + weapon.** Each boss shows a template-specific headgear and a template-specific
   weapon/shield loadout, mounted on rig bones (`head`, `handslot.r/l`) from data, with clip aliases
   still working (idle/walk/attack/hit/death). Attachments are removed with the node on death; the
   v501 dissolve and corpse paths still run.
4. **Ground aura.** A per-template themed ground aura (ring plus faint inner disc plus slow pulse and
   optional soft light) replaces the single red ring while the boss is alive and disappears on death.
   While a telegraph is active the existing telegraph-color override still wins, so v509 lane/decal
   readability is unchanged. The Performance quality tier drops the pulse and light but keeps the
   ring (matches v509 tier handling).
5. **Intro banner.** On arrival at a live full-health boss the client shows a non-blocking centered
   banner (`display_name` + `epithet`) with fade in/hold/fade out within the catalog duration; it
   plays once per boss entity id, never overlaps modal windows (mouse-transparent, below modals),
   and records a latched debug state (`played_count`, `last_template_id`, `visible`) for bots.
6. **Bot proof (client).** A client-bot scenario in `boss_floor_gate_lab` observes the banner for
   the first boss (`wait_boss_intro_banner` with `template_id` and title) and captures frames; a
   second capture scenario (or step block) covers the other template's seed. Total runtime stays
   within the 10 s default budget with no incidental navigation (lab world start, per CLAUDE.md
   movement-setup rule). Existing `boss_portrait_panel`, `boss_lane_telegraphs_visual`,
   `second_boss_template` (84/95) still pass.
7. **Captures per boss.** Real-renderer captures (Balanced tier) for Cave Warden and Crypt Matron:
   arena wide shot at isometric and first-person camera, close-up showing headgear/weapon, banner
   mid-fade, and a telegraph frame proving aura/lane coexistence. Performance tier capture for one boss.
   Evidence under `.artifacts/screenshots/` referenced from the as-built note.
8. **Cost.** No more than +1 draw-call group and +1 light per boss on Balanced; paired
   `process_ms`/frame-time probe on the boss lab before/after; first-spawn model instancing for the
   mage scene is not worse than the existing skeleton scenes beyond noise (v495 hitch baseline).
9. **Maintainability.** `main.gd` net line change <= 0 (logic lives in new focused files);
   `boss_visuals_controller.gd` stays <= 600 lines; no new `helpers=globals()` sites.

## Likely surfaces

- **Shared:** `shared/assets/boss_presentation.v0.json` + `.schema.json` (new);
  `shared/assets/kit_monster_presentation.v0.json` (+schema if a key is added): one new monster
  entry `monster_kit_skeleton_mage` (clip profile `kaykit_skeleton_ranged`); `shared/assets/monster_visuals.v0.json`
  + schema enum + `MonsterVisualsLoader` scene list (`monster_kit_skeleton_mage`);
  `assets/manifests/assets.v0.json` (+ `kaykit_skeleton_mage_v0`, `kaykit_skeleton_staff_v0`,
  `kaykit_skeleton_shield_large_a_v0`, optional `_b`).
- **Client assets:** `client/assets/monsters/kaykit_skeletons/skeleton_mage.glb`, `skeleton_staff.glb`,
  `skeleton_shield_large_*.glb` (+ `.import`, shared texture reuse), `client/scenes/monster_kit_skeleton_mage.tscn`.
- **Client code (new):** `boss_presentation_loader.gd` (static singleton, ensure_loaded),
  `boss_intro_banner.gd`, `boss_presentation_mounter.gd` (extra headgear/attachments on top of
  `KitMonsterVisual`), plus a procedural headgear builder if the decision below lands on primitives.
- **Client code (edit):** `boss_visuals_controller.gd` (intro trigger + effective-scale seam),
  `boss_visuals_context.gd` (narrow new callables, e.g. banner host), `boss_arena_presence.gd`
  (per-template aura), `kit_monster_visual.gd` (public `mount_attachments_extra`), `boss_health_bar.gd`
  (read-only: confirm `show_boss` title agrees with `display_name`; avoid layout edits),
  `main.gd` (<= a few lines: model override at `_make_entity_node`, effective scale in
  `_entity_visual_scale`/`_apply_entity_visual_metadata`, banner node creation, net <= 0 lines),
  `bot_step_catalog.gd` / `bot_scenario_runner.gd` / `bot_presentation_debug.gd` (new
  `wait_boss_intro_banner`), `client_smoke.sh` (register new tests).
- **Bot/tools:** new client scenario(s) under `tools/bot/scenarios/client/`, catalog +
  `docs/progress/scenario-movement-audit.tsv`, `tools/validate_boss_presentation.py` (split out of
  `validate_shared.py`, which is a cohesion hotspot) with `tools/test_validate_boss_presentation.py`.
- **Docs:** CODEMAP "Boss system" row, lifecycle row, as-built, license/origin notes.

## Adopt / borrow / reject (client art and Godot code)

- **Adopt:** `KitMonsterVisual` + `KitMonsterPresentationLoader` + `KitPieceLibrary` (bone attachment
  mounting, clip aliasing); `BossArenaPresence` (extend, do not fork); the v509 `wait_boss_health_bar`
  / `capture_frame` bot pattern; render quality tiers from `render_presentation.v0.json`.
- **Borrow:** the `kit_monster_presentation` catalog shape for attachments; `showme`/`regen-screenshots`
  capture conventions; `aura_soft_lights.gd` for the optional soft light.
- **Reject:** new asset packs or plugins (user decision); scaling by tinting the kit atlas (ADR-0018
  P3c: tint multiplies the atlas, use aura/emissive accents instead); client-owned boss identity
  derived from monster stats; editing `boss_templates.v0.json` `visual` values (would force golden and
  scenario 84/95 churn for a presentation change).
- **Headgear sub-decision (resolve in Task 1, see plan):** staged KayKit packs have no standalone
  helmet/crown prop (Knight_Helmet etc. are nodes inside the Adventurers hero GLBs). Default: Crypt
  Matron's headgear is the Skeleton_Mage hood/hat baked in the mage model; Cave Warden's headgear is
  a small code-native horn/crown primitive set mounted on the `head` bone (data-described shape,
  color, size), consistent with prior "code-native placeholder" precedent. Alternative (reject by
  default): extract `Knight_Helmet` mesh from the staged Adventurers GLB as a standalone asset.

## Focused verification

- `.venv/bin/pytest tools/test_validate_boss_presentation.py tools/test_validate_shared.py -q`; `make validate-shared`; `make validate-assets`.
- `make client-unit` (new `test_boss_presentation_loader.gd`, `test_boss_intro_banner.gd`, extended
  boss arena presence/ visuals tests) or targeted `GODOT ... --script res://tests/<file>`.
- Protocol regression: `make bot scenario=second_boss_template`, `make bot scenario=boss_floor_gate`
  (wire unchanged).
- Client bot: `make bot-client SCENARIO=boss_intro_banner HEADLESS=1` (state-only), then non-headless
  capture runs; also `boss_portrait_panel` and `boss_lane_telegraphs_visual`.
- `make maintainability` (file-size and extraction-coupling ratchets); no `lint-determinism` (no game/ edits).
- Not run per slice: `make ci`, `make ci-full` (coordinator owns the combined gate).

## Dependencies and integration risks

- **v511 (monster variants / `kit_monster_presentation`):** same catalog JSON/schema/`monster_visuals`
  enum/`MonsterScenesByVisual`/manifest; additive on both sides, textual conflicts likely. See report.
- **v513 (monster anim polish):** edits `kit_monster_visual.gd`, clip profiles, `animation_controller.gd`;
  v512 only adds a method and one new monster entry using an existing profile.
- **v514-v517 UI:** v517 touches `boss_health_bar.gd`/HUD; v512 keeps those files read-only except
  possibly a one-line title source change, and adds the banner as an independent CanvasLayer node.
- `main.gd` is grandfathered (baseline 6632; currently 6645); edits must net <= 0 or trigger an
  extraction.
- Model override interacts with `rec["visual_scale"]` consumers (`BossArenaPresence`,
  `BossLaneMarker`, telegraph marker). Introduce one effective-scale accessor and route them all
  through it; unit-test the three consumers.
- Boss entity may be sent only inside an interest radius: banner then fires on first sight rather
  than on stair arrival. Acceptable, documented in as-built.
- Kit atlas tint: existing `_apply_model_tint` multiplies the atlas; use a modest `tint`/none for
  bosses and carry identity in aura + attachments.

## Open questions (affect planning)

1. Scale ownership: client catalog `scale` overrides the wire `visual_scale` (default, no golden churn)
   versus retuning `boss_templates.v0.json` `visual.scale` (single source, but touches
   `boss_floor_-5.json` golden and scenarios 84/95). Default: client override.
2. Cave Warden headgear: code-native horns/crown (default) or extract the Knight helmet mesh from the
   Adventurers pack?
3. Banner: epithet text authored by us (e.g. "Warden of the Deep Stone", "Matron of the Crypt")
   acceptable? Any audio sting is out of scope.
4. Should the stale `model_pool` and `visual.model` in `boss_templates.v0.json` be removed in a
   follow-up slice after this lands?

## Implementation notes (as-built deviations)

- Crypt Matron headgear is the baked-in Skeleton_Mage hat (`headgear.kind: "none"`); a primitive crown
  was hidden inside the hat. Cave Warden gets a gold primitive crown. Both skeleton heads sit ~0.9 m
  above the `head` bone, so headgear needs the new optional `headgear.lift` (head-bone-local meters).
- Model scenes are `monster_kit_skeleton_warden` (warrior rig, no baked attachments) and
  `monster_kit_skeleton_mage`; the boss catalog mounts axe + large shield / staff. No
  `monster_visuals`, schema enum, or `MonsterScenesByVisual` edits were needed (catalog loads
  `res://scenes/<visual_key>.tscn` itself), which removes those v511 conflict surfaces.
- `kit_monster_visual.gd` is unchanged; mounting lives in `boss_presentation_mounter.gd`.
- Client scenarios are `111_boss_intro_banner_cave_warden.json` and `112_boss_intro_banner_crypt_matron.json`.
  The chest-view capture was dropped (the player overlaps the boss in the lab).
- AC8 "+1 draw-call group": the aura adds ring + disc (+2 draws) and one light on Balanced; the kit
  bodies/weapons themselves cost 12-20 draw calls per boss vs 2 for the legacy primitive skeleton.

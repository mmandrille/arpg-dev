# v512 Plan — Boss Presence

- **Spec:** [`v512_spec-boss-presence.md`](../specs/v512_spec-boss-presence.md)
- **Baseline commit:** `425b9ae4` (detached batch worktree, removed; evidence preserved under `.artifacts/batch-v511-v517-evidence/v512-boss-presence/`))
- **Prerequisites:** none unintegrated (v509 already on main). Tasks 1-4 and 8 (schema, validator,
  loader, assets) can start immediately. Siblings v511/v513 share files; re-diff against the integrated
  state before Task 5 and Task 9 (see conflicts).
- **Final gate:** focused slice verification only. The coordinator runs the combined `make ci` after
  all accepted slices integrate. No per-slice `make ci` / `make ci-full`, no commit/push.

## Spec review gate (result)

- Scope/non-goals: presentation only; no server, protocol, golden, or gameplay-rule change (ADR-0001 D2 holds).
- Determinism/replay: no `game/` edits; `lint-determinism` not required. Replay unaffected.
- Data ownership: presentation in `shared/assets/boss_presentation.v0.json`; no tuning value hardcoded
  in GDScript (scale, colors, radii, durations, bones, asset ids all catalog-owned). No
  Go/GDScript/Python code-owned tuning exceptions needed.
- Acceptance-to-test map: AC1 -> validator + pytest; AC2/3 -> loader + mounter unit tests, capture;
  AC4 -> arena presence unit test + capture; AC5 -> banner unit test + client bot; AC6 -> scenarios;
  AC7 -> captures; AC8 -> perf probe; AC9 -> `make maintainability`.
- Minor correction recorded: the brief listed `kit_monster_visual.gd` as a surface; the model for
  Crypt Matron must change from legacy `monster_skeleton` to a kit mage, which requires one new
  `kit_monster_presentation` entry and manifest assets (additive).
- Open questions in the spec have safe defaults; none block Tasks 1-4. Confirm Q1/Q2 with the user
  before Task 5/6 if they want a different default.

## File map and ownership

New (owned by v512): `shared/assets/boss_presentation.v0.json`, `.schema.json`;
`tools/validate_boss_presentation.py`, `tools/test_validate_boss_presentation.py`;
`client/scripts/boss_presentation_loader.gd`, `boss_intro_banner.gd`, `boss_presentation_mounter.gd`;
`client/tests/test_boss_presentation_loader.gd`, `test_boss_intro_banner.gd`;
`client/scenes/monster_kit_skeleton_mage.tscn`; `client/assets/monsters/kaykit_skeletons/skeleton_{mage,staff,shield_large_a}.glb`(+import);
client scenario `tools/bot/scenarios/client/111_boss_intro_banner.json` (verify number free at integration);
`docs/as-built/v512_boss-presence.md`.

Edited, shared with siblings (conflict-prone): `shared/assets/kit_monster_presentation.v0.json` (+schema),
`shared/assets/monster_visuals.v0.json` (+schema enum), `client/scripts/monster_visuals_loader.gd`,
`client/scripts/main.gd` (`MonsterScenesByVisual`, `_make_entity_node`, `_entity_visual_scale`, `_apply_entity_visual_metadata`),
`assets/manifests/assets.v0.json`, `client/scripts/kit_monster_visual.gd`, `client/scripts/boss_visuals_controller.gd`,
`client/scripts/boss_visuals_context.gd`, `client/scripts/boss_arena_presence.gd`, `client/scripts/bot_step_catalog.gd`,
`client/scripts/bot_scenario_runner.gd`, `client/scripts/bot_presentation_debug.gd`, `scripts/client_smoke.sh`,
`tools/validate_shared.py` (one import + call only), `docs/CODEMAP.md`, `docs/progress/scenario-catalog.md`,
`docs/progress/scenario-movement-audit.tsv`, `.maintainability/*` baselines only if a grandfathered file shrinks.

Read-only unless forced: `boss_health_bar.gd`, `boss_lane_marker.gd`, all `server/`, `shared/rules/`, `shared/golden/`.

## Tasks

- [x] **1. Asset decision + stage kit assets.** Confirm headgear decision (code-native horns/crown for
  Cave Warden; mage hat for Crypt Matron). Pack `Skeleton_Mage.glb`, `Skeleton_Staff.gltf`,
  `Skeleton_Shield_Large_A.gltf` from `/Users/mmandrille/git/arpg-dev/.artifacts/kaykit` (read-only source) into
  `client/assets/monsters/kaykit_skeletons/` with `tools/assets/gltf_to_glb.py`; add manifest entries
  (`kaykit_skeleton_mage_v0`, `kaykit_skeleton_staff_v0`, `kaykit_skeleton_shield_large_a_v0`) with origin/CC0
  notes and `required_nodes` (mage: `head`, `handslot.r`, `handslot.l`). Run Godot import in this worktree
  only; revert unrelated rewritten `.glb.import` files (memory: godot-import-churn).
  Check: `make validate-assets`.
- [x] **2. Catalog + schema.** Write `boss_presentation.v0.json` for `cave_warden` and `crypt_matron`:
  `display_name`, `epithet`, `visual_key`, `scale` (e.g. warden ~2.7, matron ~2.0; tune in captures, keep
  ratio >= 1.15), `tint` (optional, near-white), `attachments[]` (asset_id/bone), `headgear`
  (`kind: asset|primitive`, shape, color, size, bone), `aura` (`color`, `radius`, `inner_alpha`,
  `pulse_hz`, `light_energy`, `light_range`), `banner` (`duration_s`, `fade_in_s`, `fade_out_s`, `title_color`).
  Add `monster_kit_skeleton_mage` entry in `kit_monster_presentation.v0.json` (profile `kaykit_skeleton_ranged`, empty attachments) and the
  scene/enum registration in `monster_visuals` schema + loader + `MonsterScenesByVisual` + new `.tscn`.
  Check: `.venv/bin/python tools/validate_shared.py` (schema step only at this point).
- [x] **3. Validator + pytest (red then green).** `tools/validate_boss_presentation.py` cross-checks:
  entry per boss template, asset ids in manifest, bones in `required_nodes`, scale > 0 and pairwise ratio
  >= 1.15, banner/aura ranges, `visual_key` exists in `kit_monster_presentation`. Wire one call into
  `validate_shared.py`. Tests in `tools/test_validate_boss_presentation.py` cover each failure.
  Check: `.venv/bin/pytest tools/test_validate_boss_presentation.py tools/test_validate_shared.py -q`.
- [x] **4. Loader (`boss_presentation_loader.gd`).** `class_name ... extends RefCounted`, `static var _loaded`,
  `ensure_loaded()`, `entry(template_id)`, `effective_scale(rec)` (catalog scale else wire `visual_scale`),
  `visual_key(template_id)`. Unit test `test_boss_presentation_loader.gd` (known ids, fallback for unknown
  template, malformed JSON degrades). Register in `client_smoke.sh`.
  Check: targeted Godot test run.
- [x] **5. Model + scale override seam (main.gd, net <= 0 lines).** In `_make_entity_node`, if
  `boss_template_id` resolves to a catalog entry, use its `visual_key` instead of server `visual_model`
  (and skip the `BOSS_VISUAL_MODEL` player-scene special case for catalogued bosses). Route
  `_entity_visual_scale` and `_apply_entity_visual_metadata` through `BossPresentationLoader.effective_scale`,
  storing `rec["visual_scale"]` as effective and `rec["server_visual_scale"]` as wire so
  `BossArenaPresence` / `BossLaneMarker` / telegraph marker divisions stay consistent. Offset any added
  `main.gd` lines by moving the small boss branches into `boss_visuals_controller.gd`.
  Check: `make client-unit` subset (`test_factories`, `test_boss_health_bar`, `test_boss_lane_marker`);
  `wc -l client/scripts/main.gd` <= pre-edit count.
- [x] **6. Attachments + headgear (`boss_presentation_mounter.gd`).** Mount catalog attachments and
  headgear on bones after `KitMonsterVisual` is ready (public `mount_extra(attachments)` on
  `KitMonsterVisual`; no change to its clip aliasing). Primitive headgear built from data. Respect
  `apply_model_tint` (attachments must tint/untint consistently with telegraph flash and not leak when
  node frees). Unit test: mounter on a stub rig finds bones, skips missing ones with a warning.
  Check: targeted Godot tests + `make model model=<boss visual asset> CHECK=1` for the new mage asset.
- [x] **7. Ground aura (`boss_arena_presence.gd`).** Replace constant color/radius with catalog aura
  (ring + faint inner disc + pulse via a tween/shader-free scale-alpha oscillation; optional OmniLight3D
  via `aura_soft_lights.gd` conventions). Keep telegraph-color override precedence. Performance tier:
  ring only. Extend the existing arena-presence test: per-template color/radius, telegraph precedence,
  removal on death, tier gating.
  Check: targeted Godot test.
- [x] **8. Intro banner (`boss_intro_banner.gd`).** CanvasLayer below modal windows, mouse_filter ignore,
  title + epithet labels, fade in/hold/fade out from catalog; `play(template_id)`, latched debug state
  (`played_count`, `last_template_id`, `visible`). Trigger in `BossVisualsController.sync_boss_health_bar`
  (or a sibling method it calls): once per boss entity id, only when `hp == max_hp` and no active
  `boss_phase`; re-armed on level change. Banner host provided through `BossVisualsContext` (new narrow
  callable, no `globals()` helpers). Unit test with synthetic records: plays once, not on damaged
  reconnect, re-arms on new level, durations from catalog.
  Check: `test_boss_intro_banner.gd`.
- [x] **9. Bot step + scenario.** Add `wait_boss_intro_banner` (`template_id`, `title`, `played`/`visible`,
  `timeout_s`) to `bot_step_catalog.gd`, runner, and `bot_presentation_debug.gd` state. New client scenario
  `boss_intro_banner` in `boss_floor_gate_lab` (seed for Cave Warden as in `boss_portrait_panel`; second
  seed for Crypt Matron as in `second_boss_template`), `wait_ws_open` then `wait_boss_intro_banner`,
  `wait_entity_reaction` for `has_boss_arena_presence`, `capture_frame` (`skip_if_headless`). Lab world
  start only: no `use_stair`/`walk_to_*`/`click_floor` setup (movement-contract rule; add the scenario to
  the movement audit TSV as non-movement). Keep default 10 s budget; banner wait <= 5 s.
  Check: `make bot-client SCENARIO=boss_intro_banner HEADLESS=1`;
  `.venv/bin/pytest tools/test_scenario_movement_audit.py -q`.
- [x] **10. Regression pass.** `make bot scenario=second_boss_template`, `make bot scenario=boss_floor_gate`,
  `make bot-client SCENARIO=boss_portrait_panel HEADLESS=1`,
  `make bot-client SCENARIO=boss_lane_telegraphs_visual HEADLESS=1` (wire unchanged, health bar and lane markers intact).
- [x] **11. Real-renderer captures and cost.** Non-headless runs of the new scenario(s) for each boss:
  wide arena (isometric + first-person), close-up (headgear/weapon), banner mid-fade, telegraph +
  aura coexistence, plus one Performance-tier frame. Save under `.artifacts/screenshots/<timestamp>/` and
  preserve for handoff. Paired boss-lab frame-time/draw-call probe before/after (same host, 3 runs each);
  record any unverified limit explicitly. Iterate scale/aura values in the catalog only.
- [x] **12. Maintainability + docs.** `make maintainability` (main.gd net <= 0, no new `globals()`
  helpers, new files <= 600 lines); update `docs/CODEMAP.md` Boss system row, lifecycle row (as
  coordinator-integrated), scenario catalog, movement audit; write `docs/as-built/v512_boss-presence.md`
  (proof, captures, measured limits, scale/model-override decision, stale `model_pool` follow-up).
  Tick every checkbox in this file.

## Final focused verification (worker gate)

```bash
# run from a checkout containing the v512 changes (original batch worktree was removed)
.venv/bin/pytest tools/test_validate_boss_presentation.py tools/test_validate_shared.py tools/test_scenario_movement_audit.py -q
make validate-shared && make validate-assets && make maintainability
make client-unit                      # or targeted: boss presentation loader/banner/arena presence/factories/boss bar/lane marker
make bot scenario=second_boss_template && make bot scenario=boss_floor_gate
make bot-client SCENARIO=boss_intro_banner HEADLESS=1
make bot-client SCENARIO=boss_portrait_panel HEADLESS=1
make bot-client SCENARIO=boss_lane_telegraphs_visual HEADLESS=1
git diff --check
```

The coordinator runs the combined `make ci` only after every accepted batch slice is integrated.

## Handoff to coordinator

Report: base commit, dirty state, complete changed/deleted/untracked list (including binary `.glb` and
`.import` files and `docs/`), ignored evidence to preserve (`.artifacts/screenshots/<timestamp>/`, perf
probe logs), exact commands + outcomes, unmet criteria, and the conflict list below. Do not clean the
worktree until the coordinator verifies transfer.

## Integration / shared-file conflicts

| File | Sibling | Resolution |
|------|---------|------------|
| `shared/assets/kit_monster_presentation.v0.json` + schema | v511 (monster variants) | additive keyed entries on both sides; merge by key, re-run validator |
| `shared/assets/monster_visuals.v0.json` + schema `scene` enum, `monster_visuals_loader.gd` scene list, `main.gd` `MonsterScenesByVisual` | v511 | additive enum/dict entries; union both |
| `assets/manifests/assets.v0.json` | v511, v513 (if new clips/assets) | additive; diff final merged file vs each handoff |
| `kit_monster_visual.gd`, clip profiles, `animation_controller.gd` | v513 | v512 adds one method + uses existing profile; v513 first, rebase v512 mount call |
| `main.gd` | v511, v513, v514-v517 | keep v512 edits to 3-4 call-site lines; net <= 0 |
| `boss_health_bar.gd`, HUD layering, `boss_visuals_context.gd` | v517 | v512 adds banner as independent CanvasLayer; v517 integrates after v512 or re-verify banner z-order |
| `bot_step_catalog.gd`, `bot_scenario_runner.gd`, `bot_presentation_debug.gd` | UI slices adding bot steps | additive lists; union and recheck |
| `scripts/client_smoke.sh`, `docs/CODEMAP.md`, scenario catalog/movement audit TSV, `tools/validate_shared.py`, client scenario number 111 | all | registry-only; reconcile at integration |

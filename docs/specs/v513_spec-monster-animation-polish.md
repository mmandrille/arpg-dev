# v513 Spec — Monster Combat Animation Polish

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)
- **Date:** 2026-10-01
- **Base commit:** `425b9ae4`
- **Batch:** v511–v517 (this slice: `monster-combat-animation-polish`)
- **Dependencies:** none hard. Sibling overlap on `shared/assets/kit_monster_presentation.v0.json` with v511 (see Integration risks).

## Purpose

Monsters currently look stiff: one attack clip per rig that is **never played for ordinary melee or ranged attacks**, one hit clip, one idle played in lock-step by every copy, no spawn-in, and a windup that is only a ground ring. Use the animation clips already embedded in the KayKit skeleton rig (95 clips) and the Quaternius wolf/bat rigs to add attack variants, a windup-synchronised attack pose, spawn-in, idle variation, and direction-aware hit reactions. Everything is client-only presentation driven by events the server already emits (ADR-0007). No new art (ADR-0018 D2: existing embedded clips only).

## Findings from the code (baseline facts the plan must respect)

1. `ClientConstants.MONSTER_EVENT_CLIPS` maps only `monster_damaged -> hit` and `monster_killed -> death`. `MonsterAttackAnimationEvents.play_source_attack_for_event` plays the source monster's attack clip **only for `attack_style` `dive`/`pounce`**. So a skeleton warrior or wolf doing a plain melee hit never plays `attack`. This is the biggest visible gap.
2. `monster_attack_windup` (server `monster_melee_windup.go`) carries `source_entity_id`, `target_entity_id`, `total_ticks` (data: 8-10 ticks = 0.8-1.0 s at 10 Hz) and `attack_style` (`melee`/`pounce`). Today it only drives the v345 ground ring (`monster_melee_windup_marker.gd`). The damage event lands when the windup ends.
3. `KitMonsterVisual._alias_clips` already aliases arbitrary logical names from a profile's `clips` map (schema `clips.additionalProperties` is a string, so extra logical names such as `attack_b`, `spawn`, `hit_left` are already schema-legal). `loop` lists loop-flagged logicals.
4. `AnimationController` is a plain RefCounted with one-shot / terminal / locomotion priority and `play_one_shot(name, attack_mode, speed_scale)`. `MeleeLungePresentation` fires only for `attack_mode == "melee"` and clip `attack`/`attack_off_hand`; monsters pass no mode, so no lunge misfires.
5. `ModelReactionController.play_hit` already leans the root away from the damage source (procedural, directional) and flashes tint; `_has_death_clip()` suppresses the lean for rigged deaths. Hit-react "direction" therefore means *clip* direction (wolf has `Idle_HitReact_Left/Right`; kit has `Hit_A`/`Hit_B`), layered on the existing lean.
6. Step `[10]` of `tools/assets/validate_assets.py` checks only that a monster's `clip_profile` exists and asset ids resolve. It does **not** check that profile clip ids exist inside the GLB. `KitMonsterVisual` only `push_warning`s at runtime on a missing clip.
7. Relevant measured clip lengths (s): kit `1H_Melee_Attack_Chop` 1.07, `Slice_Diagonal` 1.00, `Slice_Horizontal` 1.07, `Stab` 1.60, `Jump_Chop` 1.33, `2H_Ranged_Shoot` 1.07, `2H_Ranged_Aiming` 1.60, `Hit_A` 0.67, `Hit_B` 0.87, `Idle` 1.07, `Idle_B` 2.13, `Idle_Combat` 4.27, `Spawn_Ground` 1.30, `Spawn_Ground_Skeletons` 3.57, `Skeletons_Awaken_Standing` 1.00. Wolf: `Attack` 1.33, `Idle` 3.33, `Idle_2` 3.33, `Idle_2_HeadLow` 4.0, `Eating` 2.5, `Idle_HitReact_Left/Right` 0.67 each, `Gallop_Jump` 0.92. Bat: `Bat_Attack` 0.88, `Bat_Attack2` 1.25, `Bat_Hit` 1.25 (long; needs speed-up), only 5 clips total.
8. `main.gd` is a grandfathered 6645-line file (baseline 6632, +25 allowed). This slice must not grow it; new logic goes in new focused files with at most a one-line call-site change.

## User-visible behavior

1. **Plain monster attacks animate.** Melee monsters play an attack clip when they strike; skeleton melee picks among several attack clips so consecutive swings differ.
2. **Windup pose.** When `monster_attack_windup` arrives for a rigged monster, its attack clip starts immediately, speed-scaled so the clip's configured *contact point* coincides with the windup end (the tick damage lands). The existing ground ring stays. Monsters with no windup play the attack clip at the damage event instead.
3. **Spawn-in.** A monster that appears via a live `entity_spawn` (not part of the initial snapshot, not an AOI stream-in of a long-standing monster) plays a short spawn clip, interruptible by the first movement, attack, hit, or death.
4. **Idle variation.** Idle monsters start their idle loop at an entity-derived phase offset (crowds no longer breathe in unison) and occasionally play an alternate idle clip (skeleton `Idle_B`/`Idle_Combat`, wolf `Idle_2`/`Idle_2_HeadLow`) on a data-driven interval, then return to idle.
5. **Hit-react direction.** The hit clip is chosen by which side the damage came from relative to the monster's facing (wolf left/right clips) or alternates between variants (kit `Hit_A`/`Hit_B`). The existing procedural lean/flash/hit-stop stays.
6. Death, v501 rim flash, v492 death burst, corpse lifecycle, and all gameplay timing are unchanged.

## Data ownership

Presentation tuning lives in `shared/assets/kit_monster_presentation.v0.json` clip profiles (schema-backed, additive, backward compatible). Proposed optional per-profile keys (final shape fixed in the plan, T1):

- `clips`: new logical names (`attack_b`, `attack_c`, `windup`/`spawn`, `idle_alt`, `idle_alt_b`, `hit_b`, `hit_left`, `hit_right`) aliased onto embedded clips.
- `variants`: `{ "attack": ["attack","attack_b",...], "hit": [...], "idle_alt": [...] }` selection groups.
- `attack_contact`: `{ logical_clip: contact_fraction }` for windup speed-scaling; `speed_scale_min/max` clamp.
- `hit_directional`: `{ "left": "hit_left", "right": "hit_right" }` (wolf) vs plain alternation (kit).
- `spawn`: `{ clip, enabled }`; `idle_variation`: `{ min_interval_s, max_interval_s, phase_offset }`.

Hardcoding any of these numbers in GDScript needs a note in the plan; the default is data. Missing/invalid optional keys must degrade to today's behavior (single clips), never break the monster.

## Non-goals

- No server, protocol, schema-version, golden, replay, determinism, or balance changes. No new events. Windup ticks stay owned by `monsters.v0.json`.
- No new art, no AnimationTree/blend spaces (ADR-0007), no IK, no animation-driven gameplay timing.
- No player/companion/hero animation changes (`kit_hero_clips.gd` untouched). No boss-specific choreography (v512 owns boss presence).
- No monster look/tint/variant work (v511), no death-pose changes beyond leaving v501 intact.
- **Adopt:** embedded KayKit and Quaternius clips; existing `AnimationPlayer` alias path. **Borrow:** v345 windup event routing and v501 loader/schema pattern. **Reject:** retargeting external animation packs, Godot AnimationTree, and any Asset Library plugin.

## Acceptance criteria

1. Every clip id referenced by any `clip_profiles` entry (including new keys) exists in the profile's rigs' embedded clips. Validator step `[10]` is extended to resolve them from the GLB and fails on an unknown id. `make validate-assets` and `make validate-shared` pass.
2. A headless client unit test (new, per CLAUDE.md rule 9, pure data/RefCounted, no `main.gd` scene paths) covers: profile loading/back-compat when optional keys are absent; every catalog scene aliases every referenced clip; attack and hit variant selection is deterministic for a given entity id + counter and spreads across variants; windup speed-scale math (contact fraction, clamps, zero/short windup); directional hit side from facing vs source; idle-variation interval bounds and phase offset derived from entity id; spawn cancel on movement/attack/hit/death; terminal state suppresses all new one-shots.
3. Existing event-driven animation tests still pass unchanged: `test_animation.gd`, `test_death_pose_ownership.gd`, `test_kit_monsters.gd`, `test_monster_death_presentation.gd`, `test_melee_lunge_presentation.gd`, `test_attack_animation_scaling.gd`.
4. Plain melee `attack_style: melee` strikes play an attack clip on the source monster; `pounce`/`dive` keep their existing clips. Ranged monsters play their ranged attack clip if a source-attributed trigger event exists; otherwise ranged stays unchanged and the plan records the deferral (see Open questions).
5. On a monster with `attack_windup_ticks > 0`, a real-renderer capture shows the attack pose building during the ring, and the measured clip time at the damage event is within one tick of the configured contact fraction. Monsters without windup still swing at the damage event.
6. Spawn-in plays only for live spawn deltas; a monster present in the initial snapshot or re-sent by interest streaming does not replay it. Spawn never delays a hit/death reaction or leaves the monster stuck in a one-shot.
7. Hit reactions: left and right hits on a wolf pick different clips; kit hits vary. `ModelReactionController` hit/death state, `impact_feedback_count`, and terminal ownership are unchanged.
8. No frame-pacing regression: matched `dungeon_frame_pacing_probe` (client scenario `108`) / `sorcerer_multigroup_perf_probe` comparison against the base commit shows no meaningful frame-time increase with 30+ live monsters. Variation work must honor `entity_presentation_lod.gd` (far/off-camera monsters skip alt idles and spawn clips) and cap concurrent timers.
9. Showme/real-camera captures (windowed, isometric player camera) of: skeleton warrior attack variants, windup mid-pose, spawn-in, alt idle, wolf left/right hit. Inspect each image; headless checks do not satisfy this.
10. `make maintainability` and `make lint-determinism` are unaffected (`main.gd` and `game/` not grown).

## Likely surfaces

- Shared data: `shared/assets/kit_monster_presentation.v0.json` (+ `.schema.json`).
- Client: `kit_monster_visual.gd`, `kit_monster_presentation_loader.gd`, `animation_controller.gd` (additive: cancel/interrupt API, optional speed scale, idle hook), `model_reaction_controller.gd` (only if hit direction needs a side-sign accessor; keep lean semantics), `monster_attack_animation_events.gd` (extend beyond dive/pounce), `monster_melee_windup_marker.gd` or a new `monster_windup_pose.gd`, new `monster_anim_variants.gd` (pure selection/math, importable standalone), new `monster_idle_variation.gd`, one-line call sites in `main.gd`/`gameplay_feedback_presentation.gd`.
- Tools: `tools/assets/validate_assets.py` step `[10]` (+ `tools/assets/glb_reader.py` already reads animation names), `tools/assets/test_validate_assets.py`.
- Tests: new `client/tests/test_monster_anim_variants.gd`, extend `test_kit_monsters.gd`; register in `scripts/client_smoke.sh`.
- Showme: a focused monster-animation capture (extend the `scenes` suite or add a small suite in `tools/showme/screenshot_catalog.py`).
- Bot: one client scenario for the windup/attack-pose proof (name and CI-pack placement decided in plan; keep within the 10 s budget and the v358 movement-contract rules).
- Docs: `docs/CODEMAP.md` (Assets / Hero light & visuals rows), plan, as-built `docs/as-built/v513_monster-combat-animation-polish.md`.

## Focused verification

- `make validate-shared`, `make validate-assets`, `.venv/bin/pytest tools/assets/test_validate_assets.py -v`
- `godot --headless --path client --script res://tests/test_monster_anim_variants.gd`, `test_kit_monsters.gd`, `test_animation.gd`, `test_death_pose_ownership.gd`, `test_monster_death_presentation.gd`, `test_melee_lunge_presentation.gd`
- one focused client bot scenario (windup pose) headless, then windowed with captures
- `make regen-screenshots SUITE=<monster-anim suite>`
- matched frame-pacing probe vs base commit
- `make maintainability`, `git diff --check`

No `make ci` per slice; the coordinator runs the combined gate.

## Integration risks

- **v511 (monster-variant-looks)** will most likely edit `kit_monster_presentation.v0.json` (+ schema) to add variant looks/tints. This slice edits `clip_profiles` only and adds optional keys to the schema's `clip_profiles` branch. Textual merge should be clean if v511 stays in `monsters`, but both may touch the schema `monsters`/top-level `additionalProperties:false`; coordinator must diff the merged JSON and schema and re-run `make validate-shared` + step `[10]`. Also possible overlap in `kit_monster_visual.gd` / `monster_family_accent.gd` / `kit_monster_presentation_loader.gd`. Keep this slice's loader additions as new static functions at the end of the file.
- **v512 (boss-presence)** may touch `animation_controller.gd`/`model_reaction_controller.gd` (boss scale/aura/entrance) and `boss_visuals_controller.gd`. Boss entrances must not be double-driven by the new spawn clip: this slice exposes an opt-out (`spawn.enabled=false` per profile or a `no_spawn_clip` flag) so v512 can own boss spawn. Additive-only edits to `AnimationController`.
- `main.gd` ratchet: call-site-only edits; if more than ~10 net lines are unavoidable, extract instead of raising the baseline.
- Spawn false positives (snapshot/AOI stream-in) are the main behavioral risk; mitigated by criterion 6.

## Open questions (affect planning)

1. **Ranged monster attack trigger.** Is there a source-attributed event when a ranged monster fires (projectile entity spawn only, or an event with `source_entity_id`)? If only an `entity_spawn` of a projectile, the plan will derive the shooter from the projectile record or defer ranged attack clips. Plan T0 resolves this by reading the server event emission; no user decision needed unless it requires a server event (which this slice will not add).
2. **Spawn-in scope.** Does the user want spawn-in for every live-spawned monster or only summoned/elite-minion/boss-add spawns? Default proposed: only live `entity_spawn` deltas after initial snapshot, within camera view, capped. Needs a user yes if the broader behavior risks visual noise in dense rooms.
3. **Bat attack/hit variants.** The bat rig has only 5 clips; variation is limited to `Bat_Attack` vs `Bat_Attack2` and a speed-scaled `Bat_Hit`. Accept this limit?

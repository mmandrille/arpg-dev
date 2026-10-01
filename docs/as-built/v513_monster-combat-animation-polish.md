# v513 - Monster combat animation polish (detached worktree handoff)

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)
- **Base:** `425b9ae4`. Client-only presentation (ADR-0007); no server, protocol, golden, replay, or asset changes (existing embedded clips only).

## Behavior

- **Plain melee strikes animate.** `monster_attack_animation_events.gd` now routes `player_damaged`/`player_killed` from a monster with no `attack_style` (plain melee) to `MonsterAnimDriver.on_strike`, which plays an attack variant. Before, only `dive`/`pounce` ever played a monster attack clip.
- **Attack variants.** Skeleton melee profile rotates `1H_Melee_Attack_Chop`, `Slice_Diagonal`, `Slice_Horizontal` per entity (deterministic hash plus counter).
- **Windup pose.** `monster_attack_windup` (melee style) starts the attack clip speed-scaled so the profile's `attack_contact` fraction lands as the server windup ends (`windup_speed_scale`, clamped by `windup_speed`). The v345 ring is untouched. The strike that follows does not replay the swing; a hit does not cut into the windup pose.
- **Hit direction.** Wolf picks `Idle_HitReact_Left`/`Right` by which side of its facing the damage source is on; skeletons alternate `Hit_A`/`Hit_B`; bat plays `Bat_Hit` at 1.6x. The existing procedural lean/flash/hit-stop in `ModelReactionController` is unchanged.
- **Spawn-in.** First appearance through an `entity_spawn` delta plays `Spawn_Ground` (skeletons), interruptible by first movement, attack, hit, or death. Skipped for level-arrival batches (they carry `wall_layout_update`), off-camera monsters, dead monsters, profiles with `spawn.enabled=false` (wolf, bat), and when 4 spawn clips are already running.
- **Idle variation.** Controllers whose AnimationPlayer carries a kit profile get an entity-keyed idle phase offset and an occasional alternate idle (`Idle_B`/`Idle_Combat`, wolf `Idle_2`/`Idle_2_HeadLow`) on a 4-9 s (wolf 5-12 s) data interval, cancelled by movement, max 6 concurrent.

## Data (`shared/assets/kit_monster_presentation.v0.json`, `clip_profiles` only)

New logical clips (`attack_b/c`, `hit_b`, `hit_left/right`, `idle_alt`, `idle_alt_b`, `spawn`) plus optional `variants`, `attack_contact`, `windup_speed`, `hit_directional`, `hit_speed_scale`, `strike_on_damage`, `spawn`, `idle_variation`. Schema additions are optional keys in the `clip_profiles` branch. Absent keys fall back to the old single-clip behavior. Structural caps (max 4 spawn clips, 6 alt idles, 0.5 s windup follow-through) are constants in code, not gameplay tuning.

## Files

New: `client/scripts/monster_anim_variants.gd` (pure math), `monster_anim_driver.gd` (event routing), `monster_idle_variation.gd`, `client/tests/test_monster_anim_variants.gd` (47 checks), `tools/bot/scenarios/client/monster_windup_pose.json`.
Edited: `animation_controller.gd` (additive: `cancel_one_shot`, `mark_motion_interruptible`, `is_idle`, `has_clip`, `clip_length`, `profile`, debug `one_shot`/`speed_scale`, idle-variation attach), `kit_monster_visual.gd` (stores the profile on the AnimationPlayer meta), `monster_attack_animation_events.gd`, `main.gd` (+5 lines: preload, windup call, live-spawn call, hit routing replaces one line; 6650 in this slice worktree vs ratchet baseline 6632; 6654 on integrated `main` after v511/v512), `bot_scenario_runner.gd` (two new `wait_entity_reaction` keys), `tools/assets/validate_assets.py` step [10] (clip ids resolved against the GLB's embedded animations, dangling variant/ref check) and its tests, `scripts/client_smoke.sh`, `docs/CODEMAP.md`, `docs/progress/scenario-movement-audit.tsv`.

## Evidence

| Check | Result |
|---|---|
| `tools/validate_shared.py` (2238), `validate_codemap.py`, `tools/assets/validate_assets.py` (452, includes new clip resolution) | PASS |
| `pytest tools/assets/test_validate_assets.py tools/test_scenario_movement_audit.py` | 30 passed |
| Godot headless: `test_monster_anim_variants` (47), `test_kit_monsters`, `test_animation`, `test_death_pose_ownership`, `test_monster_death_presentation`, `test_melee_lunge_presentation`, `test_delta_apply` | PASS |
| `make bot-client SCENARIO=monster_windup_pose HEADLESS=1` and windowed | PASS (windup clip active with speed scale < 1 about 1.2 s after engaging) |
| `make maintainability`, `git diff --check` | PASS |
| Frame pacing: `dungeon_frame_pacing_probe`, 24 live monsters, Balanced 1920x1080, 3 pairs vs base `425b9ae4`, `ARPG_PERF_DEBUG=1` | Frame interval p95 after 17.59/17.37/17.28 ms vs before 17.14/17.65/17.62 ms; draw calls 108 both; p50 14.4-14.7 ms both. No regression. Process p99 spikes (about 1 s, first-spawn-adjacent) and a p95 outlier occurred on both sides; the host was shared with other worktree runs, so this is noise-limited |

Captures (ignored, `.artifacts/`): `bot-captures/v513_monster_windup_pose.png` (live player camera, windup pose beside the ring; small at this zoom), `v513/skeleton_attack_variants.png`, `v513/skeleton_spawn_idle_hit.png`, `v513/wolf_hit_and_idle.png` (real-renderer clip pose sheets from a scratch script), `v513/pace_before/report{1,2,3}.txt` and `v513/pace/after{1,2,3}/`.

## Limits and deferrals

- **Ranged monsters:** `player_damaged` arrives at projectile impact, too late to read as the shot, and no source-attributed fire event exists (no server event added per approval). Ranged profile sets `strike_on_damage: false`; ranged attack clips are deferred.
- **Wolf/pounce and bat/dive keep their existing clips;** wolf windup pose is not applied (pounce style windup would need a crouch clip that the rig lacks). Bat gets only hit speed-up.
- **Spawn-in semantics:** `entity_spawn` is also used for fog-of-war reveals, so an idle skeleton first seen through fog plays its spawn clip (cancelled on movement). Level-arrival batches are excluded. If this reads badly in play, set `spawn.enabled=false` in the profile. Boss adds also match; v512 can opt bosses out via the profile.
- Contact fractions (0.55 chop, 0.45 others, 0.4 wolf) were tuned from pose sheets, not frame-accurate; wolf left/right clip-to-side mapping is by clip name and unverified against the model's actual left.
- Idle phase keys use the AnimationPlayer instance id (not entity id), so phase is not reproducible across runs; selection math itself is deterministic and tested.
- No showme catalog suite was added (the showme runner needs a new Godot focus scene); evidence uses the bot capture and a scratch pose-sheet script instead.
- Idle-variation timers were exercised only by unit test and the pacing probe (idle crowd), not visually in motion.

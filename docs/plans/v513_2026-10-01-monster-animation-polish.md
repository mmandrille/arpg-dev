# v513 Plan — Monster Combat Animation Polish

- **Spec:** [`docs/specs/v513_spec-monster-animation-polish.md`](../specs/v513_spec-monster-animation-polish.md)
- **Date:** 2026-10-01
- **Baseline commit:** `425b9ae4` (detached batch worktree, removed; evidence preserved under `.artifacts/batch-v511-v517-evidence/v513-monster-anim-polish/`))
- **Prerequisites:** none. T0-T9 can all proceed before sibling slices integrate. Re-check drift against v511/v512 before the final handoff.
- **Final gate:** focused slice verification only. The coordinator runs the combined `make ci` after every accepted slice is integrated; there is no per-slice `make ci`/`make ci-full` here.

## Review gate (spec review result)

- Scope/non-goals: client-only; no protocol, schema-version, golden, Go, replay, or world-preset change; determinism lint untouched (`game/` not edited).
- Server authority: windup/damage timing stays server-owned; animation only follows events. Clip speed-scaling is cosmetic and clamped.
- Shared data: all new tuning in `kit_monster_presentation.v0.json` (+ schema). No hardcoded tuning in GDScript except structural defaults, noted in T1.
- Asset adopt/borrow/reject recorded in spec (existing embedded clips only; no new files in `client/assets/`).
- Maintainability: `main.gd` (6645, baseline 6632) must end <= baseline: call-site-only edits. `model_reaction_controller.gd` (414) and others are under 600; new files stay < 600 lines. No `helpers=globals()`.
- Spec correction recorded: the spec's finding 1 (plain melee attacks never play `attack`) is the real baseline gap; the plan treats it as in scope.
- Sibling overlap: see "Shared-file conflicts" below.

## File map and ownership

| File | Change | Owner note |
|------|--------|------------|
| `shared/assets/kit_monster_presentation.v0.json` | add logical clips, `variants`, `attack_contact`, `hit_directional`, `spawn`, `idle_variation` to `clip_profiles` only | **shared with v511** |
| `shared/assets/kit_monster_presentation.v0.schema.json` | additive optional properties inside `clip_profiles` | **shared with v511** |
| `client/scripts/kit_monster_presentation_loader.gd` | new static accessors appended at end (`variants`, `attack_contact`, ...) | shared with v511 |
| `client/scripts/kit_monster_visual.gd` | alias all logical clips (already generic); expose resolved profile on the node (`meta` or getter) | shared with v511 |
| `client/scripts/monster_anim_variants.gd` (new) | pure static selection/math: variant pick, windup speed scale, hit side | importable standalone |
| `client/scripts/monster_idle_variation.gd` (new) | idle phase offset + alt-idle scheduling | |
| `client/scripts/monster_windup_pose.gd` (new) | windup event -> attack clip scaled to contact | |
| `client/scripts/animation_controller.gd` | additive: `cancel_one_shot(name)`, `current_one_shot()`, idle-return hook; behavior unchanged for existing callers | **possible v512 overlap** |
| `client/scripts/monster_attack_animation_events.gd` | play attack variant for `melee` style too | |
| `client/scripts/gameplay_feedback_presentation.gd` / `model_reaction_controller.gd` | hit clip selection call; keep lean/flash/state | |
| `client/scripts/main.gd` | call sites only (windup event, entity_spawn, hit); net lines <= 0 vs baseline | ratchet |
| `tools/assets/validate_assets.py`, `tools/assets/test_validate_assets.py` | step `[10]` resolves clip ids against GLB animation names | |
| `client/tests/test_monster_anim_variants.gd` (new), `client/tests/test_kit_monsters.gd`, `scripts/client_smoke.sh` | tests + registration | |
| `tools/showme/screenshot_catalog.py` (+ showme scene) | monster-animation capture set | |
| `tools/bot/scenarios/client/<new>_monster_windup_pose.json` + `docs/progress/scenario-catalog.md` + movement audit if needed | windup proof | v358 rules |
| `docs/CODEMAP.md`, `docs/as-built/v513_monster-combat-animation-polish.md` | docs | |

## Tasks

### T0 - Discovery (read-only, no edits)
- [x] Confirm the event emitted when a plain melee monster strikes (`player_damaged` with `source_entity_id` + `attack_style`?) in `server/internal/game/sim.go` / `monster_*` and in `shared/protocol/state_delta.v8.schema.json`.
- [x] Find the ranged-monster fire trigger (monster `attack_mode: ranged`, `tools/bot/scenarios/25_ranged_monster_ai.json`): source-attributed event vs projectile `entity_spawn`. Record result; if no clean source, defer ranged attack clips and say so in the as-built.
- [x] Determine whether dungeon monsters arrive as `entity_spawn` deltas on interest streaming vs initial snapshot, and whether summoned/minion spawns are distinguishable (spawn reason/field). Decide the spawn guard (spec criterion 6); if not distinguishable, narrow spawn-in to "after initial snapshot + in camera view + not previously seen id".
- [x] Dump clip lengths for any new clip used (use `tools/assets/glb_reader.py`) into the as-built table; the spec lists the known ones.
- Check: written notes in the as-built draft; no code changes.

### T1 - Data + schema + validator (smallest runnable: validators)
- [x] Extend `clip_profiles` for `kaykit_skeleton_melee` (attack variants Chop/Slice_Diagonal/Slice_Horizontal; `hit_b`=`Hit_B`; `idle_alt`=`Idle_B`, `idle_alt_b`=`Idle_Combat`; `spawn`=`Spawn_Ground`; `attack_contact` fractions per clip), `kaykit_skeleton_ranged` (alt idle, spawn, hit_b; ranged attack unchanged), `quaternius_wolf` (`hit_left`/`hit_right`, `idle_alt`=`Idle_2`, `idle_alt_b`=`Idle_2_HeadLow`, optional `Eating` rejected as too long/off-theme unless captured well), `quaternius_bat` (`attack_b`=`Bat_Attack2` only if it reads as a normal attack, else leave dive-only; speed-scaled hit). Add `hit_directional` (wolf), `spawn.enabled` (opt-out for v512 bosses), `idle_variation` interval bounds.
- [x] Update the schema additively (all new keys optional; keep `additionalProperties:false` valid).
- [x] Extend `validate_assets.py` step `[10]`: for each profile clip id (all logical names, including new keys and `variants` members), resolve against the animation names of every monster asset using that profile (reader on `runtime_path`); fail with profile/logical/clip in the message. Add pytest cases (unknown clip fails, valid passes, back-compat profile without new keys passes).
- Checks: `make validate-shared`, `make validate-assets`, `.venv/bin/pytest tools/assets/test_validate_assets.py -v`.

### T2 - Pure selection/math module + unit test (TDD first)
- [x] Write `client/tests/test_monster_anim_variants.gd` first (see spec criterion 2), register in `scripts/client_smoke.sh`.
- [x] Implement `monster_anim_variants.gd`: `pick_variant(group, entity_id, counter)` deterministic (hash of id + counter, no RNG singleton), `windup_speed_scale(clip_len_s, contact_fraction, windup_ticks, tick_s, min, max)`, `hit_side(monster_forward, source_dir)` -> left/right, with safe fallbacks (missing keys -> base logical clip).
- Check: `godot --headless --path client --script res://tests/test_monster_anim_variants.gd`.

### T3 - Alias + loader plumbing
- [x] Loader accessors for the new optional keys (static, appended; `duplicate(true)` like existing).
- [x] `KitMonsterVisual` aliases every logical in `clips` (already generic; ensure `loop` flags apply, non-listed are one-shot); expose the resolved profile to callers.
- [x] Extend `test_kit_monsters.gd`: every catalog scene aliases every referenced clip; back-compat when optional keys absent.
- Checks: `test_kit_monsters.gd`, `test_animation.gd`.

### T4 - Attack variants for plain melee (+ no-windup monsters)
- [x] Extend `monster_attack_animation_events.gd` so `melee` attack events (not only `dive`/`pounce`) play the picked attack variant on the source monster, skipping when a windup pose already owns that attack (T5 flag on the record) and when the monster is terminal.
- [x] Wire the existing call sites unchanged; no new `main.gd` lines if possible.
- Checks: new unit cases; `test_melee_lunge_presentation.gd` (no lunge for monster clips), `test_animation.gd`.

### T5 - Windup pose
- [x] `monster_windup_pose.gd`: on `monster_attack_windup`, resolve the monster's profile, pick the attack variant, compute speed scale from `total_ticks` and `attack_contact`, `play_one_shot(variant, "", scale)`; mark the record so T4 does not replay at the damage event; clear on death/cancel. Keep `MonsterMeleeWindupMarker` ring untouched.
- [x] One-line call next to the existing `MonsterMeleeWindupMarkerScript.sync_from_event` in `main.gd`.
- [x] `AnimationController`: restore `speed_scale` to 1.0 on cancel/terminal (already does on finish).
- Checks: unit cases (scale math, clamps, zero ticks, terminal); bot scenario in T9.

### T6 - Hit-react direction
- [x] On `monster_damaged` with hit contact, pick the hit clip: `hit_left`/`hit_right` via `hit_side` for profiles with `hit_directional`, else alternate variants; play through `AnimationController.play_one_shot` while `ModelReactionController.play_hit` continues unchanged. Do not interrupt an in-flight windup/attack clip with a hit unless the profile marks hit as interrupting (default: hit does not cancel a windup pose, matching the player rule at `main.gd:2256`).
- Checks: unit cases; `test_death_pose_ownership.gd`; `test_animation.gd`.

### T7 - Spawn-in
- [x] Apply the T0 guard; play `spawn` one-shot on a new live monster; cancel on first movement/attack/hit/death via the new `AnimationController.cancel_one_shot`; respect `spawn.enabled` and LOD (`entity_presentation_lod.gd`): off-camera/far monsters skip it.
- [x] Verify no first-spawn hitch regression (v495): spawn clip is aliased at `_ready` like all clips; no extra resource loads at spawn.
- Checks: unit cases (snapshot/known-id does not replay, cancel paths); `test_kit_monsters.gd`.

### T8 - Idle variation
- [x] `monster_idle_variation.gd`: seeded phase offset on first idle (`seek` from entity id), and a per-controller one-shot timer (`SceneTreeTimer`, bounded concurrency, skipped under LOD) that plays an alt idle when still idle and not terminal/one-shot, then returns to idle.
- Checks: unit cases (interval bounds from data, phase offset deterministic, no alt while moving/attacking/terminal).

### T9 - Bot scenario and real-camera proof
- [x] Add a client bot scenario (new lab-world setup per v358 movement contract: no incidental navigation; use a lab world / `click_entity`) that places a windup-capable monster (see `attack_windup_ticks` 8-10 in `shared/rules/monsters.v0.json`), waits for the windup event, asserts the monster's `controller` debug state shows the attack clip with scale < 1 during windup, and the clip position near the contact fraction at the damage event. Register in `docs/progress/scenario-catalog.md`; keep within 10 s.
- [x] Showme/regen-screenshots capture set (skeleton attack variants x2, windup mid-pose, spawn mid-clip, alt idle, wolf left/right hit). Showme catalog suite DEFERRED (needs a new Godot showme focus); replaced by a scratch pose-sheet capture, see as-built.
- Checks: `make bot-client SCENARIO=<new> HEADLESS=1`; windowed `AUTOPLAY_STEP_DELAY=0.05 make bot-visual scenario=<new>` with inspected capture; `make regen-screenshots SUITE=<suite>`; inspect every PNG.

### T10 - Frame pacing
- [x] Matched comparison against base `425b9ae4` with `make bot-client SCENARIO=dungeon_frame_pacing_probe` / `sorcerer_multigroup_perf_probe` (same host, same scenario; read `tools/bot/dungeon_frame_report.py` output). Record frame-time and `d_evt`/`process_ms` numbers. If a regression appears, reduce timer/alt-idle counts via LOD before accepting.

### T11 - Docs, maintainability, handoff
- [x] Update `docs/CODEMAP.md` (Assets and Hero-light/visuals rows: new scripts/tests/data).
- [x] Write `docs/as-built/v513_monster-combat-animation-polish.md`: what shipped, clip table, ranged/bat limits, evidence paths, deferrals.
- [x] `make maintainability`; verify `main.gd` <= baseline; new files < 600 lines; no `helpers=globals()`.
- [x] `git diff --check`.
- [x] Handoff to coordinator: base commit, dirty status, full changed/deleted/untracked list, ignored evidence (capture PNGs under `.artifacts/`), exact commands and outcomes, measured limits, unmet criteria, and shared-file conflicts.

## Final focused verification (no `make ci`)

```
make validate-shared && make validate-assets
.venv/bin/pytest tools/assets/test_validate_assets.py tools/test_regen_screenshots.py -v
godot --headless --path client --script res://tests/test_monster_anim_variants.gd
godot --headless --path client --script res://tests/test_kit_monsters.gd
godot --headless --path client --script res://tests/test_animation.gd
godot --headless --path client --script res://tests/test_death_pose_ownership.gd
godot --headless --path client --script res://tests/test_monster_death_presentation.gd
godot --headless --path client --script res://tests/test_melee_lunge_presentation.gd
make bot-client SCENARIO=<new windup scenario> HEADLESS=1
make maintainability && git diff --check
```

The coordinator runs combined `make ci` only after all accepted slices are integrated.

## Shared-file conflicts and integration notes

- `shared/assets/kit_monster_presentation.v0.json` and `.schema.json`: v513 edits `clip_profiles` + schema `clip_profiles` branch; v511 likely edits `monsters`/look data. Integrate the coordinator-chosen first slice, then three-way merge the other; diff the merged JSON against both handoffs and re-run `make validate-shared`, `make validate-assets`.
- `client/scripts/kit_monster_presentation_loader.gd`, `kit_monster_visual.gd`: append-only additions here to keep merges textual.
- `client/scripts/animation_controller.gd`, `model_reaction_controller.gd`: possible v512 overlap; v513 changes are additive. Spawn clip must not double-drive boss entrances (`spawn.enabled` opt-out).
- `docs/CODEMAP.md`, `docs/progress/scenario-catalog.md`, `scripts/client_smoke.sh`, `tools/showme/screenshot_catalog.py`: registry rows, reconcile by hand.
- `main.gd`: call sites only; ratchet-sensitive across the whole batch since several slices may add lines.

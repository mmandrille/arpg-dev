# v461 Plan — Entity Locomotion Polish

Status: Ready for implementation
Goal: Make ordinary dungeon traversal read as smoother grounded walking for the local hero and
nearby moving monsters without changing authoritative gameplay movement.
Architecture: Keep all gameplay authority on the server and treat this as a client presentation
slice. Reuse the existing local visual smoothing (`MovementVisualSmoothing`), remote entity tick
smoothing (`EntityTickSmoothingRuntime`), and isometric camera follow damping
(`PlayerCameraController`) as the three movement-feel layers to tune together. Treat the in-flight
v458 camera follow fix as code reality that this slice must preserve and verify; do not duplicate
that slice as a separate implementation track.
Tech stack: Godot client, shared JSON presentation catalogs, client-bot visual scenarios, docs.

## Baseline and shortcut decision

- Builds on v299 (`movement-acceleration-smoothing`), v349 (`movement-tick-smoothing`), v367
  (`camera-follow-damping`), v368 (`remote-adaptive-smoothing`), and the code already landed from
  v458 (`camera-follow-smoothing-fix`).
- Adopt: existing in-repo camera, smoothing, and bot-visual infrastructure.
- Borrow: current client-unit coverage and scenarios `80_movement_visual_smoothing`,
  `84_entity_tick_smoothing`, and `05_click_to_move`.
- Reject: external camera plugins, new art/assets, and server-authoritative movement rewrites.
- Shortcut decision: stay client-side and presentation-only unless perceptual review proves a tiny
  data-owned movement-speed retune is required; default is no gameplay retune.

## File map
| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/main.gd` | Keep hero reconciliation, visual smoothing, and remote entity movement integration coherent without reintroducing camera snaps or extra visual hops. |
| Modify | `client/scripts/movement_visual_smoothing.gd` | Tune or extend local anchor-offset smoothing behavior for ordinary walking corrections. |
| Modify | `client/scripts/entity_tick_smoothing_runtime.gd` | Tune remote walking interpolation behavior for monsters/companions/remote players. |
| Modify | `client/scripts/player_camera_controller.gd` | Preserve continuous isometric follow during locomotion and keep camera behavior aligned with v458. |
| Modify | `client/scripts/player_movement_feel.gd` | Only if needed for presentation-owned local walking feel calculations. |
| Modify | `shared/assets/movement_presentation.v0.json` | Own any new or adjusted remote entity smoothing tuning values. |
| Modify | `shared/assets/camera_presentations.v0.json` | Own any adjusted isometric follow damping values if tuning changes are needed. |
| Modify | `shared/assets/combat_feel_presentation.v0.json` | Only if an existing presentation-owned local smoothing field is the right owner for final hero tuning. |
| Modify | `client/tests/test_camera_mode_settings.gd` | Lock camera follow continuity and v458 behavior. |
| Modify | `client/tests/test_entity_tick_smoothing.gd` | Lock monster/remote smoothing duration and settle behavior. |
| Modify | `client/tests/test_movement_visual_smoothing.gd` | Lock local hero smoothing activation and settle behavior. |
| Modify | `tools/bot/scenarios/client/80_movement_visual_smoothing.json` | Preserve hero visual proof if assertions or route need minor tightening. |
| Modify | `tools/bot/scenarios/client/84_entity_tick_smoothing.json` | Tighten remote entity locomotion proof; extend if current path is too weak perceptually. |
| Modify | `tools/bot/scenarios/client/25_ranged_monster_ai.json` | Reuse the existing monster movement lab as the focused visible monster locomotion proof if a tighter route or timing window is needed. |
| Modify | `PROGRESS.md` | Advance current status and note any deferred movement retune or broader locomotion follow-up. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v461 completion row. |
| Create | `docs/as-built/v461_entity-locomotion-polish.md` | Record what the slice proved and what remained deferred. |
| Modify | `docs/specs/v461_spec-entity-locomotion-polish.md` | Mark status complete if the slice ships as planned. |

## Maintenance ratchet
Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [ ] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Defer a broad coordinator extraction with rationale: this slice should keep `main.gd` edits
  narrowly scoped to locomotion integration. If implementation pressure would grow `main.gd`, do a
  focused extraction of the new locomotion glue rather than advancing the baseline again.

Verification:
```bash
make maintainability
```

## Task 1 — Confirm locomotion baseline and tune ownership
Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/movement_visual_smoothing.gd`
- Modify: `client/scripts/entity_tick_smoothing_runtime.gd`
- Modify: `client/scripts/player_camera_controller.gd`
- Modify: `shared/assets/movement_presentation.v0.json`
- Modify: `shared/assets/camera_presentations.v0.json`
- Modify: `shared/assets/combat_feel_presentation.v0.json` (only if needed)

- [x] Step 1.1: Confirm the v458 camera-follow fix behavior already present in code is the active
  baseline and document any missing integration this slice must preserve.
```bash
rg -n "_reconcile_player|tick_follow\\(|follow_damping_seconds|remote_adaptive|movement_visual_smoothing" client/scripts/main.gd client/scripts/player_camera_controller.gd client/scripts/entity_tick_smoothing_runtime.gd client/scripts/movement_visual_smoothing.gd shared/assets/camera_presentations.v0.json shared/assets/movement_presentation.v0.json
```

- [x] Step 1.2: Adjust local hero smoothing and/or camera-follow interplay so ordinary
  click-to-move corrections no longer read as back-and-forth snapping.
```bash
godot --headless --path client --script res://tests/test_movement_visual_smoothing.gd
godot --headless --path client --script res://tests/test_camera_mode_settings.gd
```

- [x] Step 1.3: Adjust remote entity smoothing tuning and/or duration logic so nearby monsters read
  as walking during chase without over-smoothing large discontinuities.
```bash
godot --headless --path client --script res://tests/test_entity_tick_smoothing.gd
make validate-shared
```

## Task 2 — Lock focused unit coverage
Files:
- Modify: `client/tests/test_camera_mode_settings.gd`
- Modify: `client/tests/test_entity_tick_smoothing.gd`
- Modify: `client/tests/test_movement_visual_smoothing.gd`

- [x] Step 2.1: Add or tighten camera assertions so isometric follow continuity remains preserved
  across routine locomotion updates and explicit snap behavior stays limited to setup/discontinuity
  paths.
```bash
godot --headless --path client --script res://tests/test_camera_mode_settings.gd
```

- [x] Step 2.2: Add or tighten entity smoothing assertions for monster/remote segment duration,
  activation, and settle behavior using rule-derived or semantic expectations.
```bash
godot --headless --path client --script res://tests/test_entity_tick_smoothing.gd
```

- [x] Step 2.3: Add or tighten hero visual smoothing assertions so ordinary walking corrections
  activate and settle cleanly without changing gameplay anchors.
```bash
godot --headless --path client --script res://tests/test_movement_visual_smoothing.gd
```

## Task 3 — Bot and visual proof
Files:
- Modify: `tools/bot/scenarios/client/80_movement_visual_smoothing.json`
- Modify: `tools/bot/scenarios/client/84_entity_tick_smoothing.json`
- Modify: `tools/bot/scenarios/client/05_click_to_move.json` only if required by the final proof
- Modify: `tools/bot/scenarios/client/25_ranged_monster_ai.json` only if the existing monster
  locomotion proof needs a tighter route or timing window

- [x] Step 3.1: Re-run and, if needed, tighten `80_movement_visual_smoothing` so it still proves
  hero-side smoothing activation and settle under the final tuning.
```bash
HEADLESS=1 make bot-visual scenario=80_movement_visual_smoothing
```

- [x] Step 3.2: Re-run and, if needed, extend `84_entity_tick_smoothing` so it proves nearby remote
  locomotion continuity strongly enough for this slice.
```bash
HEADLESS=1 make bot-visual scenario=84_entity_tick_smoothing
```

- [x] Step 3.3: Keep a normal gameplay sanity proof for ordinary traversal and add one focused
  visible monster locomotion proof using the existing ranged-monster chase lab if the smoothing
  change is not obvious enough in `84_entity_tick_smoothing`.
```bash
HEADLESS=1 make bot-visual scenario=05_click_to_move
HEADLESS=1 make bot-visual scenario=25_ranged_monster_ai
```

## Task 4 — Lifecycle docs and close-out
Files:
- Modify: `docs/specs/v461_spec-entity-locomotion-polish.md`
- Create: `docs/as-built/v461_entity-locomotion-polish.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark the spec complete and write the as-built summary with the exact locomotion
  behaviors proven and any deferred follow-ups.
```bash
git diff -- docs/specs/v461_spec-entity-locomotion-polish.md docs/as-built/v461_entity-locomotion-polish.md
```

- [x] Step 4.2: Update lifecycle and progress tracking for v461, including any deferred broader
  movement-speed retune or monster locomotion follow-up.
```bash
git diff -- PROGRESS.md docs/progress/slice-lifecycle.md
```

## Final verification
- [x] `make validate-shared`
- [x] `make maintainability`
- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-visual scenario=80_movement_visual_smoothing`
- [x] `HEADLESS=1 make bot-visual scenario=84_entity_tick_smoothing`
- [x] `HEADLESS=1 make bot-visual scenario=05_click_to_move`
- [x] `HEADLESS=1 make bot-visual scenario=25_ranged_monster_ai`
- [ ] `make ci`

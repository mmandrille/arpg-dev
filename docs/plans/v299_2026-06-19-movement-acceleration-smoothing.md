# v299 Plan — Movement Acceleration Smoothing

Status: Complete
Goal: Smooth only the local `CharacterVisual` child during small anchor movement steps while
preserving exact gameplay anchor positions.
Architecture: Add a focused client helper that stores the prior anchor position, offsets the
visual child opposite small anchor steps, and eases the child back to zero. `player_anchor`,
`predicted_pos`, camera target, picking, reach, and all server messages remain unchanged.
Tech stack: Godot client presentation, Godot unit tests, Godot client bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v298 and the existing local `PlayerAnchor/CharacterVisual` scene hierarchy. Asset/plugin
decision: adopt the in-repo hierarchy and bot debug-state patterns; borrow click-to-move and
attack-move scenarios; reject external plugins/assets and gameplay-position smoothing. Presentation
constants stay in the helper because they are local visual feel constants, not gameplay balance.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/movement_visual_smoothing.gd` | Preserve visual continuity for small anchor moves and ease the visual child back to zero. |
| Modify | `client/scripts/main.gd` | Wire helper into local player anchor updates and bot debug state without changing gameplay position. |
| Create | `client/tests/test_movement_visual_smoothing.gd` | Cover small-step offset, catch-up, reset, and debug state. |
| Modify | `client/scripts/bot_step_catalog.gd` | Register smoothing wait/assert step names. |
| Modify | `client/scripts/bot_wait_handlers.gd` | Delegate smoothing wait matching. |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Assert smoothing debug state. |
| Create | `tools/bot/scenarios/client/80_movement_visual_smoothing.json` | Live client proof that smoothing activates after movement and settles. |
| Modify | `docs/specs/v299_spec-movement-acceleration-smoothing.md` | Mark complete when shipped. |
| Create | `docs/as-built/v299_movement-acceleration-smoothing.md` | Summarize shipped behavior and limits. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v299 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and next selected slice. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines, and touched grandfathered files stay
inside allowance.

Hotspot files:
- [x] `client/scripts/main.gd` (grandfathered; stayed at 5811 lines, within the 5813 allowance)
- [x] Other touched files are under 600 lines.

Decision:
- [x] Put smoothing behavior in a new focused helper.
- [x] Keep `main.gd` integration to narrow state/call sites and avoid smoothing gameplay anchors.

Verification:

```bash
make maintainability
```

## Task 1 — Smoothing Helper

- [x] Step 1.1: Add helper for small-step visual offset preservation and catch-up.
- [x] Step 1.2: Reset instead of offsetting for large movement deltas.
- [x] Step 1.3: Add debug state for active/offset length.
- [x] Step 1.4: Add focused Godot unit tests.

```bash
godot --headless --path client --script res://tests/test_movement_visual_smoothing.gd
```

## Task 2 — Client Integration And Bot Assertion

- [x] Step 2.1: Wire helper into local player anchor reconciliation/prediction paths.
- [x] Step 2.2: Expose smoothing debug state from `get_bot_state`.
- [x] Step 2.3: Add bot wait/assert matching for smoothing state.
- [x] Step 2.4: Add `80_movement_visual_smoothing`.

```bash
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=80_movement_visual_smoothing HEADLESS=1 ./scripts/bot_client_local.sh
```

## Task 3 — Focused Regression Proof

- [x] Step 3.1: Rerun click-to-move and attack-move scenarios.
- [x] Step 3.2: Run maintainability.

```bash
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=05_click_to_move HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Manual visual verification:

```bash
make bot-visual scenario=80_movement_visual_smoothing
```

## Task 4 — Lifecycle Docs

- [x] Step 4.1: Mark spec and plan complete after focused checks pass.
- [x] Step 4.2: Add as-built and lifecycle row.
- [x] Step 4.3: Update `PROGRESS.md`.

## Final verification

- [x] `godot --headless --path client --script res://tests/test_movement_visual_smoothing.gd`
- [x] `godot --headless --path client --script res://tests/test_client_bot.gd`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=80_movement_visual_smoothing HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=05_click_to_move HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18083 BASE_URL=http://localhost:18083 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `make maintainability`

Autoloop batch mode: do not run `make ci` for this individual slice unless focused verification is
insufficient. The enclosing autoloop must run one final `make ci` after the selected feature queue is
committed.

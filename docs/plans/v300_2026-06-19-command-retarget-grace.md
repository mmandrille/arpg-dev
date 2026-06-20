# v300 Plan — Command Retarget Grace

Status: Complete
Goal: Queue only the latest rapid floor retarget while the local send throttle is cooling down.
Architecture: Add a focused client helper that owns one short-lived pending floor retarget, replacement
counts, expiry, and debug state. `main.gd` delegates floor movement dispatch through a small wrapper
that queues during local throttle and sends the latest command when legal. Server movement,
protocol, pathfinding, and authoritative positions remain unchanged.
Tech stack: Godot client presentation/input, Godot unit tests, Godot client bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v296 attack-move/sticky targeting and v299 movement visual smoothing. Asset/plugin
decision: adopt existing `move_to_intent`, local click cooldown, and bot debug-state patterns;
borrow v296/v299 regression scenarios; reject external plugins/assets, server command queues,
schema changes, and gameplay-position smoothing. The grace duration stays in the helper as a local
input-feel constant rather than shared gameplay balance.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/command_retarget_grace.gd` | Store one latest pending floor retarget, replace it during grace, expire stale commands, and expose debug state. |
| Modify | `client/scripts/main.gd` | Route floor click movement through retarget grace and tick dispatch after local cooldown clears. |
| Modify | `client/scripts/bot_facade.gd` | Route bot `move_to_intent` floor clicks through the same retarget-aware path. |
| Create | `client/tests/test_command_retarget_grace.gd` | Cover latest-wins replacement, dispatch readiness, and expiry. |
| Modify | `client/scripts/bot_step_catalog.gd` | Register retarget-grace wait/assert step names. |
| Modify | `client/scripts/bot_wait_handlers.gd` | Delegate retarget-grace wait matching. |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Assert retarget-grace debug state. |
| Create | `tools/bot/scenarios/client/81_command_retarget_grace.json` | Live client proof that rapid floor clicks dispatch the latest queued retarget. |
| Modify | `docs/specs/v300_spec-command-retarget-grace.md` | Mark complete when shipped. |
| Create | `docs/as-built/v300_command-retarget-grace.md` | Summarize shipped behavior and limits. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v300 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and next selected slice. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines, and touched grandfathered files stay
inside allowance.

Hotspot files:
- [x] `client/scripts/main.gd` (grandfathered; stayed at 5813 lines)
- [x] Other touched files are under 600 lines.

Decision:
- [x] Put retarget grace behavior in a new focused helper.
- [x] Keep `main.gd` integration to narrow call sites and debug state.

Verification:

```bash
make maintainability
```

## Task 1 — Retarget Grace Helper

- [x] Step 1.1: Add helper for one latest floor retarget with replacement counts and expiry.
- [x] Step 1.2: Add dispatch readiness/pop behavior when local cooldown clears.
- [x] Step 1.3: Add debug state for active/latest/replaced/dispatched/expired state.
- [x] Step 1.4: Add focused Godot unit tests.

```bash
godot --headless --path client --script res://tests/test_command_retarget_grace.gd
```

## Task 2 — Client Integration And Bot Assertion

- [x] Step 2.1: Route local floor click movement through a retarget-aware wrapper.
- [x] Step 2.2: Tick queued retarget dispatch after local cooldown clears.
- [x] Step 2.3: Expose retarget debug state from `get_bot_state`.
- [x] Step 2.4: Route bot `click_floor` through the same wrapper and add wait/assert matching.
- [x] Step 2.5: Add `81_command_retarget_grace`.

```bash
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=81_command_retarget_grace HEADLESS=1 ./scripts/bot_client_local.sh
```

## Task 3 — Focused Regression Proof

- [x] Step 3.1: Rerun attack-move and movement-smoothing scenarios.
- [x] Step 3.2: Run maintainability.

```bash
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=80_movement_visual_smoothing HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Manual visual verification:

```bash
make bot-visual scenario=81_command_retarget_grace
```

## Task 4 — Lifecycle Docs

- [x] Step 4.1: Mark spec and plan complete after focused checks pass.
- [x] Step 4.2: Add as-built and lifecycle row.
- [x] Step 4.3: Update `PROGRESS.md`.

## Final verification

- [x] `godot --headless --path client --script res://tests/test_command_retarget_grace.gd`
- [x] `godot --headless --path client --script res://tests/test_client_bot.gd`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=81_command_retarget_grace HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18083 BASE_URL=http://localhost:18083 SCENARIO=80_movement_visual_smoothing HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `make maintainability`

Autoloop batch mode: do not run `make ci` for this individual slice unless focused verification is
insufficient. The enclosing autoloop must run one final `make ci` after the selected feature queue is
committed.

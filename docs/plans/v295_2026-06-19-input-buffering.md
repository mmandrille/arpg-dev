# v295 Plan — Input Buffering

Status: Complete
Goal: Queue short-lived basic attack clicks when the local client is just outside cooldown/range,
then send the existing authoritative `action_intent` once legal.
Architecture: Keep the server authoritative and do not change the wire format. The client stores a
small presentation/input buffer for monster targets only, validates target liveness and local reach
before dispatch, and reuses the current click attack send path. Extract the buffer policy into a
focused helper so `client/scripts/main.gd` does not grow.
Tech stack: Godot GDScript client input, Godot headless tests, Godot client bot scenario, lifecycle
docs.

## Baseline and shortcut decision

Builds on v294 full-CI green. The slice adopts existing local cooldown/recovery UI, sustained click
state, local reach helpers, and `action_intent`; borrows the client bot click-to-kill scenario
style; rejects external assets/plugins, protocol changes, and server combat changes.

The v294 review/refactor cadence remains due after the selected Movement / Combat Fluidity feature
batch completes and passes final CI, per the autoloop post-loop handoff.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/combat_input_buffer.gd` | Focused helper for short-lived attack buffer state and expiry rules. |
| Modify | `client/scripts/main.gd` | Integrate queued monster clicks into current click/repeat dispatch paths without changing protocol. |
| Modify | `client/tests/test_sustained_input.gd` | Add headless helper coverage for attack-buffer replacement, expiry, and clearing. |
| Modify | `client/scripts/bot_step_catalog.gd` | Add focused wait/assertion step only if needed by the scenario proof. |
| Modify | `client/scripts/bot_wait_handlers.gd` | Add focused wait helper only if needed by the scenario proof. |
| Create | `tools/bot/scenarios/client/77_input_buffering.json` | Prove a click during local recovery still produces a later authoritative combat event. |
| Modify | `docs/specs/v295_spec-input-buffering.md` | Mark complete when shipped. |
| Create | `docs/as-built/v295_input-buffering.md` | Summarize shipped behavior and limits. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v295 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and preserve review/refactor handoff. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `client/scripts/bot_scenario_runner.gd` only if scenario validation requires it
- [x] Did every touched grandfathered file stay inside its ratchet allowance?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.
- [ ] Defer extraction with rationale: N/A

Verification:

```bash
make maintainability
```

## Task 1 — Attack Buffer Helper

Files:
- Create: `client/scripts/combat_input_buffer.gd`
- Modify: `client/tests/test_sustained_input.gd`

- [x] Step 1.1: Add a small `CombatInputBuffer` helper that stores only monster attack targets,
  expiry seconds, and replacement/clear semantics.
- [x] Step 1.2: Add headless tests for queue, replacement, expiry, clear, and target validation
  decisions without requiring a scene tree.

```bash
godot --headless --path client --script res://tests/test_sustained_input.gd
```

## Task 2 — Client Input Integration

Files:
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Preload and instantiate the helper; clear it on teardown, blocked input, player
  death, force-stand, and non-monster click commands.
- [x] Step 2.2: When a monster click arrives during local cooldown or outside local reach, queue
  that target instead of returning.
- [x] Step 2.3: Each input tick, dispatch the queued target only when the target is still a living
  monster, local cooldown is ready, and local range is legal; reuse current facing, animation,
  `_send_action_intent`, cooldown, and recovery UI logic.
- [x] Step 2.4: Keep sustained hold behavior intact; the buffer is a one-shot fallback, not a
  replacement for held attack repeats.

```bash
godot --headless --path client --script res://tests/test_sustained_input.gd
```

## Task 3 — Client Bot Proof

Files:
- Create: `tools/bot/scenarios/client/77_input_buffering.json`
- Modify: `client/scripts/bot_step_catalog.gd` only if needed
- Modify: `client/scripts/bot_wait_handlers.gd` only if needed
- Modify: `client/tests/test_client_bot.gd` only if scenario validation changes

- [x] Step 3.1: Add a focused client scenario that equips the training bow, clicks a monster, clicks
  it again during local recovery, and waits for a later authoritative combat event.
- [x] Step 3.2: Add only the minimal bot wait/assertion support needed to observe the buffered
  behavior, preferring existing `click_entity`, `wait_event`, `wait_ticks`, and
  `click_entity_until_event` primitives first.

```bash
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=01_click_to_kill HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=05_click_to_move HEADLESS=1 ./scripts/bot_client_local.sh
```

Manual visual verification:

```bash
make bot-visual scenario=77_input_buffering
```

## Task 4 — Lifecycle Docs And Focused Verification

Files:
- Modify: `docs/specs/v295_spec-input-buffering.md`
- Create: `docs/as-built/v295_input-buffering.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v295_2026-06-19-input-buffering.md`

- [x] Step 4.1: Mark the spec and plan complete after focused checks pass.
- [x] Step 4.2: Add the as-built summary and lifecycle row.
- [x] Step 4.3: Update `PROGRESS.md` current status, CI gate, next slice, and review/refactor
  handoff note.

```bash
make maintainability
```

## Final verification

- [x] `godot --headless --path client --script res://tests/test_sustained_input.gd`
- [x] `godot --headless --path client --script res://tests/test_client_bot.gd`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=01_click_to_kill HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=05_click_to_move HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `make maintainability`

Autoloop batch mode: do not run `make ci` for this individual slice unless focused verification is
insufficient. The enclosing autoloop must run one final `make ci` after the selected feature queue is
committed.

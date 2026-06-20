# v298 Plan — Hit Stop / Impact Flash

Status: Complete
Goal: Add client-only impact flash/hold feedback to existing authoritative model reactions.
Architecture: Extend the focused `ModelReactionController` that already owns hit/death visual
reactions. No server, shared rules, protocol, or `main.gd` changes are required.
Tech stack: Godot client presentation, Godot client bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v297 local attack presentation and the existing authoritative hit/death reaction mapping.
Asset/plugin decision: adopt the existing reaction controller and material-tint primitives; borrow
the combat-control client bot setup; reject new assets/plugins and global pause. The micro hold and
flash constants stay in client code because they are local presentation constants colocated with the
existing reaction constants, not gameplay balance.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/model_reaction_controller.gd` | Add immediate impact flash, tiny visual hold, and debug count for hit/death reactions. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Let existing reaction assertions require a minimum impact-feedback count. |
| Modify | `client/tests/test_animation.gd` | Assert hit flash/count while preserving restore/death behavior. |
| Create | `tools/bot/scenarios/client/79_hit_stop_impact_flash.json` | Live client proof for authoritative hit triggering impact feedback. |
| Modify | `docs/specs/v298_spec-hit-stop-impact-flash.md` | Mark complete when shipped. |
| Create | `docs/as-built/v298_hit-stop-impact-flash.md` | Summarize shipped behavior and limits. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v298 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and next selected slice. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines, and touched grandfathered files stay
inside allowance.

Hotspot files:
- [ ] `client/scripts/bot_scenario_runner.gd` (grandfathered; only a tiny matcher addition fits in current allowance)
- [ ] `client/tests/test_animation.gd` (under 600 before slice)

Decision:
- [x] Extend existing focused reaction controller instead of touching `main.gd`.
- [x] Avoid new bot runner abstractions; add only the minimum matcher predicate.

Verification:

```bash
make maintainability
```

## Task 1 — Reaction Impact Feedback

- [x] Step 1.1: Add immediate impact flash and tiny visual hold to `play_hit`.
- [x] Step 1.2: Add the same impact feedback before death lean in `enter_death`.
- [x] Step 1.3: Expose a debug `impact_feedback_count`.
- [x] Step 1.4: Add focused animation tests for count/flash and existing restore/death behavior.

```bash
godot --headless --path client --script res://tests/test_animation.gd
```

## Task 2 — Bot Assertion And Scenario

- [x] Step 2.1: Extend `wait_entity_reaction` / `assert_entity_reaction` matching with
  `impact_feedback_min`.
- [x] Step 2.2: Add `79_hit_stop_impact_flash` using the combat-control lab, training bow, and an
  authoritative damage event.

```bash
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=79_hit_stop_impact_flash HEADLESS=1 ./scripts/bot_client_local.sh
```

## Task 3 — Focused Regression Proof

- [x] Step 3.1: Rerun v295 and v296 client combat scenarios.
- [x] Step 3.2: Run maintainability.

```bash
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Manual visual verification:

```bash
make bot-visual scenario=79_hit_stop_impact_flash
```

## Task 4 — Lifecycle Docs

- [x] Step 4.1: Mark spec and plan complete after focused checks pass.
- [x] Step 4.2: Add as-built and lifecycle row.
- [x] Step 4.3: Update `PROGRESS.md`.

## Final verification

- [x] `godot --headless --path client --script res://tests/test_animation.gd`
- [x] `godot --headless --path client --script res://tests/test_client_bot.gd`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=79_hit_stop_impact_flash HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `make maintainability`

Autoloop batch mode: do not run `make ci` for this individual slice unless focused verification is
insufficient. The enclosing autoloop must run one final `make ci` after the selected feature queue is
committed.

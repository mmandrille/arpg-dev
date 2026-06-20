# v296 Plan — Attack Move / Sticky Targeting

Status: Complete
Goal: Make an out-of-range monster click move the player toward attack range and attack that same
target once the existing local reach/cooldown checks allow it.
Architecture: Client-only command memory and approach movement. The client sends existing
`move_to_intent` and `action_intent` messages; the server remains authoritative for movement,
range, cooldown, hit, damage, and death.
Tech stack: Godot GDScript client input, Godot headless tests, Godot client bot scenario, lifecycle
docs.

## Baseline and shortcut decision

Builds on v295 input buffering. The slice adopts existing movement/attack intents, local reach
constants, and bot wait helpers; borrows the combat-control lab; rejects external assets/plugins,
protocol changes, and server behavior changes.

The v294 review/refactor cadence remains due after the selected Movement / Combat Fluidity feature
batch completes and passes final CI, per the autoloop post-loop handoff.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/combat_reach.gd` | Expose local attack reach/radius and compute a monster approach point. |
| Create | `client/scripts/combat_sticky_target.gd` | Focused helper for sticky monster target state and clear guards. |
| Modify | `client/scripts/main.gd` | Start approach movement for out-of-range monster clicks and dispatch when legal. |
| Modify | `client/tests/test_sustained_input.gd` | Add helper coverage for approach-point and sticky target behavior. |
| Create | `tools/bot/scenarios/client/78_attack_move_sticky_targeting.json` | Prove far monster click moves into range and later attacks. |
| Modify | `docs/specs/v296_spec-attack-move-sticky-targeting.md` | Mark complete when shipped. |
| Create | `docs/as-built/v296_attack-move-sticky-targeting.md` | Summarize shipped behavior and limits. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v296 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and next selected slice. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines; touched grandfathered files stay inside
their ratchet allowance.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/tests/test_sustained_input.gd`

Decision:
- [x] Put sticky-target state and approach math in focused helpers.
- [ ] Defer extraction with rationale: N/A

Verification:

```bash
make maintainability
```

## Task 1 — Sticky Target And Approach Helpers

Files:
- Create: `client/scripts/combat_sticky_target.gd`
- Modify: `client/scripts/combat_reach.gd`
- Modify: `client/tests/test_sustained_input.gd`

- [x] Step 1.1: Add a helper that stores one sticky monster target and clears for missing, dead,
  non-monster, or dead-player states.
- [x] Step 1.2: Extend local reach helper with an approach point that stops just inside current
  local attack reach using existing reach/radius data.
- [x] Step 1.3: Add focused headless tests for sticky replacement/clear guards and approach-point
  geometry.

```bash
godot --headless --path client --script res://tests/test_sustained_input.gd
```

## Task 2 — Client Input Integration

Files:
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Preload and instantiate sticky target state; clear it on teardown, blocked input,
  death, force-stand, and non-monster commands.
- [x] Step 2.2: For out-of-range monster clicks, set the sticky target and send `move_to_intent` to
  the local approach point instead of only queueing a short recovery buffer.
- [x] Step 2.3: Each input tick, when sticky target is valid, local cooldown is ready, and local
  reach is legal, reuse the v295 monster dispatch path.
- [x] Step 2.4: Preserve v295 recovery buffering for in-range clicks during local cooldown and keep
  direct in-range clicks immediate.

```bash
godot --headless --path client --script res://tests/test_sustained_input.gd
```

## Task 3 — Client Bot Proof

Files:
- Create: `tools/bot/scenarios/client/78_attack_move_sticky_targeting.json`

- [x] Step 3.1: Add a scenario that clicks the control-lab dungeon mob from spawn with the
  human-like buffered click action, waits for player approach, and observes a later authoritative
  `monster_damaged` or `attack_missed` event against that target.
- [x] Step 3.2: Rerun v295 input-buffering and movement click regressions.

```bash
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=05_click_to_move HEADLESS=1 ./scripts/bot_client_local.sh
```

Manual visual verification:

```bash
make bot-visual scenario=78_attack_move_sticky_targeting
```

## Task 4 — Lifecycle Docs And Focused Verification

Files:
- Modify: `docs/specs/v296_spec-attack-move-sticky-targeting.md`
- Create: `docs/as-built/v296_attack-move-sticky-targeting.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v296_2026-06-19-attack-move-sticky-targeting.md`

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
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=05_click_to_move HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `make maintainability`

Autoloop batch mode: do not run `make ci` for this individual slice unless focused verification is
insufficient. The enclosing autoloop must run one final `make ci` after the selected feature queue is
committed.

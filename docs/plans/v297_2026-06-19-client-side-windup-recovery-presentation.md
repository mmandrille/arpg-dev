# v297 Plan — Client-Side Windup / Recovery Presentation

Status: Complete
Goal: Start local basic attack swing/audio/recovery when the client sends an attack intent, then
suppress duplicate local swing/audio when the authoritative result arrives.
Architecture: Client-only presentation tracking keyed by local attacker/target. The server remains
authoritative for all combat results and no wire format changes.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/combat_local_attack_presentation.gd` | Track a locally started basic attack target and consume the matching server result. |
| Modify | `client/scripts/main.gd` | Start local presentation on basic attack dispatch and route authoritative local-player combat events through duplicate suppression. |
| Modify | `client/scripts/bot_facade.gd` | Keep direct bot monster click presentation aligned with the main dispatch path. |
| Modify | `client/tests/test_sustained_input.gd` | Add helper tests for matching, replacement, and non-match behavior. |
| Modify | `docs/specs/v297_spec-client-side-windup-recovery-presentation.md` | Mark complete when shipped. |
| Create | `docs/as-built/v297_client-side-windup-recovery-presentation.md` | Summarize shipped behavior and limits. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v297 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and next selected slice. |

## Maintenance ratchet

Target: touched grandfathered files stay inside their ratchet allowance.

Hotspot files:
- [ ] `client/scripts/main.gd`
- [ ] `client/tests/test_sustained_input.gd`

Decision:
- [x] Put presentation tracking in a focused helper.
- [x] Offset `main.gd` integration by simplifying the duplicated local combat-event presentation path.

## Task 1 — Presentation Tracker

- [x] Step 1.1: Add a helper that records the current locally started monster attack target.
- [x] Step 1.2: Consume only matching local-player `monster_damaged`, `monster_killed`,
  `attack_missed`, or `attack_blocked` events.
- [x] Step 1.3: Add focused headless tests for replacement, match, non-match, and clear behavior.

```bash
godot --headless --path client --script res://tests/test_sustained_input.gd
```

## Task 2 — Client Presentation Integration

- [x] Step 2.1: Start local attack animation/audio/recovery from one helper when basic attack
  dispatch sends `action_intent`.
- [x] Step 2.2: Use the presentation tracker to suppress duplicate local player swing/audio on the
  matching authoritative combat result while preserving damage text and target reactions.
- [x] Step 2.3: Keep remote and monster event-driven attacks unchanged.

```bash
godot --headless --path client --script res://tests/test_client_bot.gd
```

## Task 3 — Focused Regression Proof

- [x] Step 3.1: Rerun v295 and v296 client bot combat scenarios.
- [x] Step 3.2: Run maintainability.

```bash
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Manual visual verification:

```bash
make bot-visual scenario=78_attack_move_sticky_targeting
```

## Task 4 — Lifecycle Docs

- [x] Step 4.1: Mark spec and plan complete after focused checks pass.
- [x] Step 4.2: Add as-built and lifecycle row.
- [x] Step 4.3: Update `PROGRESS.md`.

## Final verification

- [x] `godot --headless --path client --script res://tests/test_sustained_input.gd`
- [x] `godot --headless --path client --script res://tests/test_client_bot.gd`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `make maintainability`

Autoloop batch mode: do not run `make ci` for this individual slice unless focused verification is
insufficient. The enclosing autoloop must run one final `make ci` after the selected feature queue is
committed.

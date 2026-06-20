# v296 As-Built: Attack Move / Sticky Targeting

Date: 2026-06-19
Spec: [`docs/specs/v296_spec-attack-move-sticky-targeting.md`](../specs/v296_spec-attack-move-sticky-targeting.md)
Plan: [`docs/plans/v296_2026-06-19-attack-move-sticky-targeting.md`](../plans/v296_2026-06-19-attack-move-sticky-targeting.md)

## What Shipped

- Added `CombatStickyTarget`, a client-only explicit monster target memory that clears for missing,
  dead, non-monster, or dead-player states.
- Extended `CombatReach` with local approach-point calculation that uses existing weapon reach and
  target-radius data, stopping just inside local attack range.
- Updated out-of-range monster clicks to send the existing `move_to_intent` toward an approach
  point, remember the clicked monster, and later reuse the v295 monster dispatch path when local
  reach and cooldown allow it.
- Preserved immediate legal in-range attacks and v295 short recovery input buffering.
- Kept server authority intact: no protocol, server movement/pathfinding, combat legality,
  cooldown, damage, hit roll, death, loot, or monster AI changes.
- Moved bot-facing entity dispatch forwarding into the existing `BotFacade` helper to keep
  `main.gd` inside its grandfathered file-size ratchet.
- Added `78_attack_move_sticky_targeting`, which clicks the control-lab dungeon mob from spawn,
  verifies the player approaches the target, and observes a later authoritative `monster_damaged`
  or `attack_missed` event against that same target.

## Proof

Focused verification:

```bash
godot --headless --path client --script res://tests/test_sustained_input.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=05_click_to_move HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Result: green on 2026-06-19. The local bot script printed the same post-pass
`cleanup_account.py` missing-`httpx` warning in this environment, but each scenario returned
success.

## Manual Visual Command

```bash
make bot-visual scenario=78_attack_move_sticky_targeting
```

## Deferred

- Client-side windup/recovery presentation remains the next selected Movement / Combat Fluidity
  slice.
- Hit stop, movement acceleration smoothing, command retarget grace, and melee lunge remain later
  selected slices in this feature queue.
- Generic ground attack-move and auto-acquire-nearest-enemy behavior remain out of scope.

# v296 Spec: Attack Move / Sticky Targeting

Status: Complete
Date: 2026-06-19
Codename: `attack-move-sticky-targeting`
Baseline: v295 `input-buffering`

## Purpose

Make enemy clicks feel like an ARPG command: when the player clicks a living monster outside local
basic-attack reach, the client moves toward attack range, remembers that target, and sends the
existing server-authoritative `action_intent` once local reach and local cooldown allow it.

This is the second Movement / Combat Fluidity autoloop slice. It builds on v295 input buffering and
keeps the review/refactor handoff as a post-loop task for the selected feature batch.

## Non-goals

- Do not add or change protocol messages, server combat legality, pathfinding authority, cooldowns,
  reach tuning, damage, hit rolls, monster AI, or loot.
- Do not add target retarget grace, client-side windup/recovery presentation, hit stop, movement
  acceleration smoothing, or melee lunge in this slice.
- Do not implement generic ground attack-move or auto-acquire-nearest-enemy behavior; this slice is
  only sticky targeting for the explicitly clicked monster.
- Do not buffer loot, interactable, skill, directional attack, right-click skill, or floor movement
  commands.
- Do not use external assets/plugins.

## Acceptance Criteria

- Clicking a living monster outside local basic-attack reach sends a movement command toward a local
  attack-range approach point instead of doing nothing.
- The clicked monster remains the sticky attack target while the player moves; when the target is a
  living monster, local cooldown is ready, and local reach is legal, the client sends the existing
  `action_intent`.
- Clicking a different living monster replaces the sticky target and updates the approach movement.
- A floor click, loot/interactable click, force-stand command, blocked input state, player death,
  missing target, dead target, or non-monster target clears the sticky target.
- Legal in-range monster clicks continue to attack immediately through the v295 dispatch path.
- Existing v295 input buffering still works for clicks during local recovery.
- The server remains authoritative for final movement, range legality, hit/miss/damage/death, and
  all outcomes.

## Scope And Likely Files

- Client helpers: extend `client/scripts/combat_reach.gd` or add a focused helper for local
  approach-point computation and sticky target state.
- Client input: `client/scripts/main.gd`, with net line count kept inside the grandfathered
  maintainability ratchet.
- Client tests: `client/tests/test_sustained_input.gd` or another focused headless helper test for
  approach-point calculation / sticky target guards.
- Client bot proof: add `tools/bot/scenarios/client/78_attack_move_sticky_targeting.json`.
- Docs: v296 plan, as-built, lifecycle, and `PROGRESS.md`.

## Test And Bot Proof

Focused checks:

```bash
godot --headless --path client --script res://tests/test_sustained_input.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=05_click_to_move HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Manual visual verification command:

```bash
make bot-visual scenario=78_attack_move_sticky_targeting
```

## Asset And Plugin Decision

- Adopt: existing `move_to_intent`, `action_intent`, local reach constants, v295 input-buffer
  dispatch path, and client bot scenario primitives.
- Borrow: existing combat-control lab and v295 target-event wait support.
- Reject: external assets/plugins, server-side combat/pathfinding changes, protocol schema changes,
  and new visual effects for this slice.

## ADR Alignment

- ADR-0001 D2/D3: client prediction/input feel may improve, but the server owns movement and combat
  outcomes.
- ADR-0007: no animation state crosses the wire; this slice only chooses when existing intents are
  sent.
- ADR-0014 D7: target click forgiveness improves input feel without changing damage, challenge, or
  progression.

## Open Questions And Risks

- No blocking questions. Use the smallest explicit-target approach behavior that can be proven in
  the control lab.
- `client/scripts/main.gd` is already at its grandfathered file-size allowance, so implementation
  must keep net line count inside the ratchet by putting state/math in focused helpers and trimming
  touched coordinator code.

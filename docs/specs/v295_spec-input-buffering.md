# v295 Spec: Input Buffering

Status: Complete
Date: 2026-06-19
Codename: `input-buffering`
Baseline: v294 `full-ci-residual-stabilization`

## Purpose

Make basic combat clicks feel less brittle by remembering a short-lived monster attack click when
the player clicks slightly before the local attack cooldown is ready or just before they step into
local attack range. The client retries the buffered target only when it is still legal to send the
existing server-authoritative `action_intent`; the server remains the source of truth for range,
cooldown, hit, damage, death, and loot.

This is the first Movement / Combat Fluidity autoloop slice. The repo-wide review/refactor handoff
is due after v294 and is recorded as the post-loop handoff for the selected feature batch.

## Non-goals

- Do not add a protocol/schema version or a new intent type.
- Do not change server combat legality, attack cooldown formulas, reach tuning, damage, or monster
  AI.
- Do not implement attack-move pathing, sticky retargeting, command retarget grace, immediate
  windup presentation, hit stop, movement smoothing, or melee lunge; those remain later selected
  slices in this movement/combat queue.
- Do not buffer loot, interactable, skill, directional attack, right-click skill, or floor movement
  commands in this slice.
- Do not use external assets/plugins.

## Acceptance Criteria

- Clicking a living monster while the local basic-attack cooldown is still active stores that target
  briefly instead of dropping the click.
- Clicking a living monster while out of local weapon reach stores that target briefly and sends the
  existing `action_intent` once the player is inside local reach, if the buffer has not expired.
- A buffered attack is cleared when the target disappears, dies, becomes non-monster, the player
  dies, input becomes blocked, force-stand starts, or the player issues a non-monster click command.
- A newer monster click replaces the previous buffered target.
- Buffered dispatch uses the same `_send_action_intent`, local facing, attack animation,
  local cooldown, and recovery UI path as current legal monster clicks.
- Buffer timing is client-side presentation/input policy, not gameplay balance. Any duration or
  epsilon introduced by this slice is held in a focused client helper and not copied into unrelated
  tests.
- Existing click-to-kill and click-to-move behavior continues to pass.

## Scope And Likely Files

- Client input: `client/scripts/main.gd`, plus a focused helper such as
  `client/scripts/combat_input_buffer.gd` to avoid growing the large coordinator.
- Client tests: `client/tests/test_sustained_input.gd` or a new focused headless helper test for
  the attack buffer.
- Client bot proof: add or update a focused Godot client scenario under
  `tools/bot/scenarios/client/` to prove a click made during recovery still produces a later
  authoritative `monster_damaged` or `monster_killed` event.
- Docs: v295 plan, as-built, lifecycle, and `PROGRESS.md`.

## Test And Bot Proof

Focused checks:

```bash
godot --headless --path client --script res://tests/test_sustained_input.gd
godot --headless --path client --script res://tests/test_client_bot.gd
make bot-client scenario=77_input_buffering HEADLESS=1
make bot-client scenario=01_click_to_kill HEADLESS=1
make bot-client scenario=05_click_to_move HEADLESS=1
make maintainability
```

Manual visual verification command:

```bash
make bot-visual scenario=77_input_buffering
```

## Asset And Plugin Decision

- Adopt: existing client local cooldown/recovery UI, sustained-click state, monster local reach
  helpers, and `action_intent`.
- Borrow: existing client bot click-to-kill scenario style for proof.
- Reject: external assets/plugins, server-side combat changes, protocol schema changes, and new
  visual effects for this slice.

## ADR Alignment

- ADR-0001 D2/D3: the client may improve input feel and prediction, but the server owns combat
  outcomes.
- ADR-0007: no animation state crosses the wire; this slice only decides when the client sends the
  existing intent and reuses current client-only presentation.
- ADR-0014 D7: better input forgiveness reduces accidental failed recovery windows without changing
  combat challenge or damage.

## Open Questions And Risks

- No blocking questions. Default to the smallest client-only buffer that is long enough to catch a
  near-cooldown or near-range click and short enough to avoid surprising delayed attacks.
- `client/scripts/main.gd` is over the maintainability target and already at its grandfathered
  allowance, so the plan must include a small helper extraction or net line reduction.

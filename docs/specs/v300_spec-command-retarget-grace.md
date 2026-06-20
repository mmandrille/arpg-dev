# v300 Spec: Command Retarget Grace

Status: Complete
Date: 2026-06-19
Codename: `command-retarget-grace`
Baseline: v299 `movement-acceleration-smoothing`

## Purpose

Make rapid floor click changes feel deliberate instead of dropped when the local client send throttle
is still cooling down from a previous click. The player should be able to adjust the current move
destination immediately, with the client briefly preserving only the latest requested destination
and sending it as soon as the short local throttle clears.

This is the sixth Movement / Combat Fluidity autoloop slice. It builds on the click-to-move,
attack-move, and movement smoothing slices without changing server authority.

## Non-goals

- Do not change server movement, pathfinding, collision, authoritative positions, movement intent
  schema, action intent schema, or replay contracts.
- Do not queue arbitrary old commands. The grace stores only the latest local floor retarget and
  expires quickly if a longer combat recovery still blocks dispatch.
- Do not change basic attack cooldowns, attack buffering rules, sticky targeting reach logic,
  interactable activation, skill targeting, or WASD movement.
- Do not add UI, assets, plugins, animation clips, or production VFX.

## Acceptance Criteria

- A floor click that arrives during the local click/send throttle queues a short-lived retarget
  instead of being dropped.
- Additional floor clicks during that grace replace the queued destination, so the latest click wins.
- When the local throttle clears before grace expiry, the client sends one `move_to_intent` for the
  latest queued floor destination.
- If a longer cooldown outlives the grace window, the queued movement retarget expires without
  dispatching stale movement.
- Bot state exposes retarget-grace debug state, and a client bot scenario proves rapid floor clicks
  dispatch the latest retarget and move toward the final clicked destination.
- v296 attack-move/sticky targeting and v299 movement smoothing still pass.

## Scope And Likely Files

- Client helper: `client/scripts/command_retarget_grace.gd`.
- Client integration: `client/scripts/main.gd`, kept within its file-size ratchet.
- Client tests: `client/tests/test_command_retarget_grace.gd`.
- Bot assertion plumbing: `client/scripts/bot_step_catalog.gd`, `client/scripts/bot_wait_handlers.gd`,
  `client/scripts/bot_assertion_handlers.gd`.
- Bot direct-action facade: `client/scripts/bot_facade.gd`, so `click_floor` goes through the same
  retarget grace path as local floor clicks.
- Bot proof: new `tools/bot/scenarios/client/81_command_retarget_grace.json`, plus
  `78_attack_move_sticky_targeting` and `80_movement_visual_smoothing` regressions.
- Docs: v300 plan, as-built, lifecycle, and `PROGRESS.md`.

## Test And Bot Proof

Focused checks:

```bash
godot --headless --path client --script res://tests/test_command_retarget_grace.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=81_command_retarget_grace HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=80_movement_visual_smoothing HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Manual visual verification command:

```bash
make bot-visual scenario=81_command_retarget_grace
```

## Asset, Plugin, And Tuning Decision

- Adopt: existing `move_to_intent`, local click cooldown, bot debug-state, and client bot scenario
  patterns.
- Borrow: v296 attack-move and v299 smoothing regressions.
- Reject: external plugins/assets, server-side command queues, schema changes, and gameplay-position
  smoothing.
- Tuning note: the grace duration is a local input-feel constant owned by the helper because it only
  affects whether a short-lived client retarget survives the local send throttle.

## ADR Alignment

- ADR-0001 D2/D3: the server remains authoritative for movement outcomes; this slice changes only
  client input dispatch timing.
- ADR-0007: no new animation or server event is required.

## Open Questions And Risks

- No blocking questions. The helper intentionally stores one latest floor destination rather than a
  command list, avoiding stale queued movement and keeping behavior deterministic from the server's
  perspective.

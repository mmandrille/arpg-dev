# v299 Spec: Movement Acceleration Smoothing

Status: Complete
Date: 2026-06-19
Codename: `movement-acceleration-smoothing`
Baseline: v298 `hit-stop-impact-flash`

## Purpose

Make local movement starts, stops, and small reconciliation steps feel less robotic by smoothing only
the local character visual back toward the exact `PlayerAnchor`.

This is the fifth Movement / Combat Fluidity autoloop slice. It builds on v295-v298 combat feel
improvements while keeping movement authority unchanged.

## Non-goals

- Do not change server movement, pathfinding, collision, authoritative positions, prediction math,
  movement intents, combat reach, picking, camera follow target, or bot/player position state.
- Do not smooth monsters, remote players, projectiles, leap/charge motion, teleport-sized movement,
  or dungeon/object placement in this slice.
- Do not add command retarget grace or melee lunge in this slice.
- Do not add new assets, plugins, or animation clips.

## Acceptance Criteria

- `player_anchor.position` and `predicted_pos` remain exact gameplay positions.
- `CharacterVisual.position` preserves visual world continuity for small local anchor moves by
  offsetting opposite the anchor step, then eases back toward zero.
- Large spawn/teleport/correction moves reset the visual offset instead of showing misleading lag.
- Bot state exposes smoothing debug state, and a client bot scenario proves the offset activates
  after a floor move and settles afterward.
- v296 attack-move/sticky targeting still passes.

## Scope And Likely Files

- Client helper: `client/scripts/movement_visual_smoothing.gd`.
- Client integration: `client/scripts/main.gd`, kept within its file-size ratchet.
- Client tests: `client/tests/test_movement_visual_smoothing.gd`.
- Bot assertion plumbing: `client/scripts/bot_step_catalog.gd`, `client/scripts/bot_wait_handlers.gd`,
  `client/scripts/bot_assertion_handlers.gd`.
- Bot proof: new `tools/bot/scenarios/client/80_movement_visual_smoothing.json`, plus `05_click_to_move`
  and `78_attack_move_sticky_targeting` regressions.
- Docs: v299 plan, as-built, lifecycle, and `PROGRESS.md`.

## Test And Bot Proof

Focused checks:

```bash
godot --headless --path client --script res://tests/test_movement_visual_smoothing.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=80_movement_visual_smoothing HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=05_click_to_move HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Manual visual verification command:

```bash
make bot-visual scenario=80_movement_visual_smoothing
```

## Asset, Plugin, And Tuning Decision

- Adopt: existing `PlayerAnchor` / `CharacterVisual` hierarchy and bot state patterns.
- Borrow: `05_click_to_move` and `78_attack_move_sticky_targeting` regression proof.
- Reject: external plugins/assets, server movement changes, and smoothing gameplay anchors.
- Tuning note: this slice adds client-code-owned micro presentation constants for local visual
  catch-up speed and reset distance. They are presentation-only and colocated with the helper that
  owns the visual offset; a future broader movement-presentation catalog can move them to data.

## ADR Alignment

- ADR-0001 D2/D3: server remains authoritative for movement and combat outcomes.
- ADR-0007: local movement smoothing is client-only presentation and does not cross the wire.

## Open Questions And Risks

- No blocking questions. The helper resets on large movement deltas to avoid confusing spawn,
  teleport, leap, or reconciliation jumps with smooth local locomotion.

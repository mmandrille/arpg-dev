# Spec: `boss-phase-readability`

Status: Accepted
Date: 2026-06-10
Codename: `boss-phase-readability`
Slice: v57 - boss phase readability
Baseline: v56 `monster-attack-cadence`

## Purpose

The first boss already emits authoritative `boss_phase_started` / `boss_phase_ended` events and the
client already shows a persistent boss health bar. The current presentation only tints the boss
during telegraph, which makes the timing window and danger zone too subtle.

This slice improves boss readability without changing combat authority: the boss health bar shows
the active phase and an approximate countdown, and the world renders a simple warning marker during
telegraph phases using the server-provided telegraph radius/color.

## Non-goals

- No server combat, boss AI, pattern deck, damage, timing, or protocol/schema changes.
- No new boss pattern variety; that remains queued as the next selected slice.
- No client-side hit detection or gameplay decisions.
- No production VFX/audio, boss portraits, phase history, multi-boss layout, or phase-specific art.
- No external Godot plugin or asset dependency.

## Acceptance criteria

1. The boss health bar keeps existing boss title, hp/max hp, and ratio behavior.
2. When a boss phase is active, the bar exposes and displays:
   - `phase_kind`
   - `pattern_id`
   - `phase_index`
   - `duration_ticks`
   - display-only `remaining_ticks`
   - `phase_ratio`
3. The display-only phase countdown starts from authoritative `duration_ticks` on
   `boss_phase_started` and decreases locally until the phase ends or is superseded.
4. The phase state clears on `boss_phase_ended`, boss death/removal, level change, and teardown.
5. During telegraph phases, the boss node has a visible marker named `BossTelegraphMarker` using
   the authoritative telegraph radius and color.
6. The telegraph marker is removed when the telegraph phase ends or a non-telegraph phase starts.
7. `get_bot_state()` exposes boss bar phase fields and boss presentation fields so headless
   scenarios can assert readability state.
8. Client bot scenario coverage reaches the first boss floor, waits for a telegraph, asserts the
   boss health bar phase state, and asserts the in-world telegraph marker is active.

## Plugin and asset adoption

Decision: reject external Godot plugins, demos, and asset packs.

Reason: this is a narrow display-only extension of existing boss health bar and primitive marker
patterns. The project already uses in-repo procedural Godot primitives for placeholder presentation,
and a plugin would add maintenance surface without improving authority, testability, or scope.

## Scope and likely files

- `client/scripts/boss_health_bar.gd` - phase label/progress/debug state.
- `client/scripts/main.gd` - phase timer seeding, local countdown, telegraph marker attach/remove,
  and bot presentation debug state.
- `client/scripts/bot_scenario_runner.gd` - boss health bar phase expectations and boss presentation
  assertions.
- `client/tests/test_boss_health_bar.gd` - phase display/debug unit coverage.
- `client/tests/test_client_bot.gd` - runner validation/matching coverage.
- `tools/bot/scenarios/client/28_boss_phase_readability.json` - focused client-bot proof.
- `docs/specs/v57_spec-boss-phase-readability.md` - this spec.
- `docs/plans/v57_2026-06-10-boss-phase-readability.md` - implementation plan.
- `PROGRESS.md` and `docs/as-built/v57_boss-phase-readability.md` - lifecycle close-out.

## Test and bot proof

- `make client-unit`
- `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- `make bot scenario=24_boss_floor_gate.json`
- `make ci`

## Open questions and risks

- Q1: Should the countdown be authoritative to the exact tick?
  - Decision: no. This slice is display-only and uses the server event duration plus local client
    time. Exact timing remains owned by server events and bot/replay proof.
- Risk: If the boss phase begins before the scenario starts waiting, the bot must be able to observe
  either active telegraph state or the next telegraph cycle. Mitigation: use generous wait timeouts
  and assert fields through `get_bot_state()`.

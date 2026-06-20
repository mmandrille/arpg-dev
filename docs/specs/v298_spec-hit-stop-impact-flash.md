# v298 Spec: Hit Stop / Impact Flash

Status: Complete
Date: 2026-06-19
Codename: `hit-stop-impact-flash`
Baseline: v297 `client-side-windup-recovery-presentation`

## Purpose

Make successful combat impacts read with more weight by adding a tiny client-only impact flash and
visual hold to the existing model reaction path for authoritative hit/death events.

This is the fourth Movement / Combat Fluidity autoloop slice. It builds on v295 input buffering,
v296 attack move/sticky targeting, and v297 immediate local attack presentation.

## Non-goals

- Do not change server combat timing, hit rolls, damage, cooldowns, health, death, loot, or
  protocol.
- Do not pause input, networking, simulation, timers, or the whole scene tree.
- Do not add camera shake, new production VFX assets, new plugins, or external art.
- Do not add movement smoothing, command retarget grace, or melee lunge in this slice.

## Acceptance Criteria

- Authoritative `monster_damaged`, `monster_killed`, `player_damaged`, and `player_killed` events
  that already trigger a model reaction also trigger a brief impact flash/hold on that model.
- Miss, block, immune, aggro, and non-combat UI events do not trigger impact feedback.
- Existing hit/death reaction state still restores/terminates as before.
- Bot debug presentation exposes an impact-feedback count so client scenarios can assert the effect
  without timing-dependent pixel checks.
- v297 attack presentation and v295/v296 combat scenarios continue to pass.

## Scope And Likely Files

- Client reaction presentation: `client/scripts/model_reaction_controller.gd`.
- Client bot assertion matching: `client/scripts/bot_scenario_runner.gd`.
- Client tests: `client/tests/test_animation.gd`.
- Bot proof: new `tools/bot/scenarios/client/79_hit_stop_impact_flash.json`.
- Docs: v298 plan, as-built, lifecycle, and `PROGRESS.md`.

## Test And Bot Proof

Focused checks:

```bash
godot --headless --path client --script res://tests/test_animation.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=79_hit_stop_impact_flash HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Manual visual verification command:

```bash
make bot-visual scenario=79_hit_stop_impact_flash
```

## Asset, Plugin, And Tuning Decision

- Adopt: existing `ModelReactionController` hit/death tween path and material-tint primitive.
- Borrow: control-lab bow setup from v295/v296 client scenarios.
- Reject: external VFX plugins, new assets, global scene-tree pause, and server/protocol changes.
- Tuning note: this slice adds client-code-owned micro presentation constants alongside the
  existing `ModelReactionController` reaction constants. They are not gameplay balance and have no
  shared/server consumer; a future broader VFX catalog can move them to content data if needed.

## ADR Alignment

- ADR-0001 D2/D3: server remains authoritative for all outcomes.
- ADR-0007: animation/reaction feedback remains client-only and driven by existing authoritative
  combat events.

## Open Questions And Risks

- No blocking questions. The live bot assertion uses an impact-feedback counter instead of exact
  frame timing so the proof is not brittle.

# v464 Spec: Combat Input Flow Polish

Status: Complete
Date: 2026-07-14
Codename: `combat-input-flow-polish`
Baseline: v463 `dungeon-surface-detail-overlays`

## Purpose

Make basic combat control feel more continuous during ordinary dungeon fights. The slice improves
the client-side path from monster click to approach, buffered attack, retarget, and local melee
feedback so players can click into melee, queue a follow-up, or change intent without the hero
appearing to stall between movement and attack.

This is a presentation and input-flow polish slice. The authoritative Go sim continues to own all
combat outcomes, cooldown acceptance, damage, death, loot, and persistence. Existing server events
remain the source of truth for confirmed combat results.

## Non-goals

- No protocol or schema version bump.
- No server-side combat, movement, collision, AI, damage, loot, HP, XP, or persistence change.
- No basic-attack cooldown rebalance or shared-rules attack-speed retuning.
- No skill, projectile, ranged-monster, or boss-pattern redesign.
- No new external plugins, imported assets, or asset pipeline changes.
- No production combat VFX/audio pass beyond reusing the existing local attack presentation hooks.

## Acceptance criteria

1. Clicking a living monster from outside melee range keeps the existing attack-move behavior, then
   dispatches a basic attack once the target becomes locally reachable without requiring a second
   perfectly timed click.
2. Clicking a monster during local attack recovery buffers at most one follow-up attack intent and
   dispatches it when local recovery clears, as long as the target is still a living monster.
3. Clicking a different living monster while an attack buffer or sticky attack is active replaces
   the prior pending combat target instead of leaving the hero stuck on stale intent.
4. Clicking the ground during the retarget grace window after combat replaces pending combat intent
   and dispatches movement when local recovery clears.
5. Local melee attack feedback remains immediate and bounded: lunge/animation starts on local attack
   presentation, clears or settles after recovery, and never changes the authoritative player
   position.
6. Existing combat input and melee-lunge client bot proofs remain green, with a new or updated proof
   covering target replacement and movement retarget behavior.

## Scope and likely files touched

- Client input flow:
  - `client/scripts/attack_move_input_coordinator.gd`
  - `client/scripts/combat_input_buffer.gd`
  - `client/scripts/combat_sticky_target.gd`
  - `client/scripts/command_retarget_grace.gd`
  - `client/scripts/combat_local_attack_presentation.gd`
  - `client/scripts/melee_lunge_presentation.gd`
- Client presentation data:
  - `shared/assets/combat_feel_presentation.v0.json`
- Client tests:
  - `client/tests/test_command_retarget_grace.gd`
  - `client/tests/test_melee_lunge_presentation.gd`
  - `client/tests/test_combat_feel_presentation_loader.gd`
  - optional focused tests for buffer/sticky target replacement if missing
- Bot proof:
  - `tools/bot/scenarios/client/77_input_buffering.json`
  - `tools/bot/scenarios/client/82_melee_lunge_micro_step.json`
  - optional new `tools/bot/scenarios/client/103_combat_input_flow_polish.json`
- Docs:
  - `docs/plans/v464_2026-07-14-combat-input-flow-polish.md`
  - `docs/as-built/v464_combat-input-flow-polish.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

## Test and bot proof

Expected focused verification:

```bash
make validate-shared
godot --headless --path client --script res://tests/test_command_retarget_grace.gd
godot --headless --path client --script res://tests/test_melee_lunge_presentation.gd
godot --headless --path client --script res://tests/test_combat_feel_presentation_loader.gd
make client-unit
HEADLESS=1 make bot-visual scenario=77_input_buffering
HEADLESS=1 make bot-visual scenario=82_melee_lunge_micro_step
```

If the implementation adds the focused retarget proof, run:

```bash
HEADLESS=1 make bot-visual scenario=103_combat_input_flow_polish
```

Manual visual verification should use:

```bash
make bot-visual scenario=82_melee_lunge_micro_step
make bot-visual scenario=103_combat_input_flow_polish
```

## Asset and plugin decision

- Adopt: existing in-repo combat feel configuration, attack buffering, sticky target, retarget grace,
  local attack presentation, melee lunge, Godot unit tests, and client bot scenario infrastructure.
- Borrow: v428 `hero-animation-feel`, v461 `entity-locomotion-polish`, and existing scenarios
  `77_input_buffering` / `82_melee_lunge_micro_step` as the closest proof patterns.
- Reject: external input frameworks, camera/combat plugins, new art assets, and production VFX/audio
  work. The slice is a code-native client control polish pass.

## Open questions and risks

1. Existing client bot steps may not expose enough debug state to prove target replacement. If so,
   the plan should add the smallest debug assertion or bot step needed, rather than widening the
   slice into a full input tooling pass.
2. The slice intentionally does not retune authoritative attack cadence. If combat still feels slow
   after this pass, a later shared-rules slice should evaluate attack interval and movement-speed
   retuning with rule-derived tests.
3. `client/scripts/main.gd` is a known coordinator hotspot. Prefer focused helper/test changes; if a
   tiny integration edit is unavoidable, keep it below the ratchet allowance and document the
   deferral rather than extracting unrelated code inside this slice.

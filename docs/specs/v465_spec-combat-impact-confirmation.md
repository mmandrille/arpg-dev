# v465 Spec — Combat Impact Confirmation

Status: Complete
Date: 2026-07-14
Codename: combat-impact-confirmation

## Purpose

Make authoritative combat results read more clearly in the Godot client. Normal hits, kills, misses,
blocks, immunes, and critical hits should produce distinct, bounded confirmation using the existing
combat outcome punch / impact feedback path, so v464's smoother input flow lands with clearer moment-to-moment
feedback.

## Non-goals

- Do not change server combat math, attack speed, cooldowns, movement speed, hit chance, block chance, or damage.
- Do not add a protocol/schema version bump unless an existing combat event cannot identify the result type.
- Do not add production audio, imported VFX assets, camera redesign, or per-class/per-skill bespoke effects.
- Do not preserve legacy absence of hit feedback as compatibility; update tests/docs to the new presentation contract.

## Acceptance criteria

- `monster_damaged` events create a visible, short-lived combat confirmation effect near the target when enemy impact feedback is enabled.
- `monster_killed` events create a stronger but still bounded confirmation distinct from a normal hit.
- `attack_missed`, `attack_blocked`, and immune/critical outcomes remain visually distinct from normal damage.
- The confirmation path respects the existing `enemy_impact_feedback` toggle.
- Matching authoritative combat results do not cause the local predicted attack animation/audio to replay.
- New behavior is covered by focused Godot unit tests and at least one Godot client bot scenario that exercises authoritative combat results in `combat_control_lab`.

## Scope and files likely touched

- Client presentation:
  - `client/scripts/combat_outcome_punch.gd`
  - `client/scripts/combat_event_presentation.gd`
  - `client/scripts/combat_local_attack_presentation.gd` only if duplicate-prevention assertions need tightening.
- Client tests:
  - `client/tests/test_combat_outcome_punch.gd`
  - `client/tests/test_impact_sparks.gd`
  - `client/tests/test_look_and_feel_polish.gd` only if camera/player feedback changes.
- Bot proof:
  - Add or update a scenario under `tools/bot/scenarios/client/` for hit/kill confirmation.
  - Update `docs/progress/scenario-movement-audit.tsv` for any new scenario.
- Docs:
  - `PROGRESS.md`
  - `docs/as-built/v465_combat-impact-confirmation.md`

## Asset/plugin decision

Adopt: existing in-repo Godot procedural presentation nodes from `combat_outcome_punch.gd` and the existing
combat lab scenario/world data.

Borrow: existing impact feedback gating, combat text routing, and bot `wait_entity_reaction` proof patterns.

Reject: external VFX assets, audio packs, Godot plugins, and new asset pipelines. This slice is a tuning and
presentation-mapping change over existing client primitives.

## Test and bot proof

- Focused Godot tests for outcome spawn rules and normal-hit/kill node creation.
- Focused Godot tests proving enemy impact feedback disabled prevents the new hit confirmation.
- `make client-unit`.
- New or updated Godot client scenario, run visually with:

```bash
HEADLESS=1 make bot-visual scenario=104_combat_impact_confirmation
```

- Final slice proof with `make ci`.

## Open questions and risks

- Critical-hit proof should stay unit-level unless current combat lab data can make critical hits deterministic without server/rule rebalance.
- Miss/block proof can remain on the existing outcome punch unit path unless adding deterministic bot setup is cheaper than changing shared combat rules.

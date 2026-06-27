# v353 As-Built — Mobility Skill Smoothing

Date: 2026-06-26  
Spec: [`docs/specs/v353_spec-mobility-skill-smoothing.md`](../specs/v353_spec-mobility-skill-smoothing.md)  
Plan: [`docs/plans/v353_2026-06-26-mobility-skill-smoothing.md`](../plans/v353_2026-06-26-mobility-skill-smoothing.md)

## Shipped behavior

- **`movement_presentation.v0.json`**: `mobility_smoothing` tuning (`enabled`,
  `teleport_duration_seconds`, `teleport_travel_duration_seconds`).
- **`MobilitySkillPresentation`**: leap/charge/teleport tweens for local and remote players on
  `skill_cast`; teleporter floor-travel reveal when position jump exceeds threshold.
- **Tick smoothing** skipped for entities with active mobility presentation.
- **Bot debug** exposes `mobility_skill_smoothing`.
- **Extended bot proof**: `86_mobility_skill_smoothing` (barbarian leap lab; waits for wall layout,
  not monsters).

## Boundaries

- Client-only presentation; no server/protocol changes.
- Earthbreaker/disengage variants out of scope.

## Verification

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_mobility_skill_presentation.gd
HEADLESS=1 make bot-client SCENARIO=86_mobility_skill_smoothing
```

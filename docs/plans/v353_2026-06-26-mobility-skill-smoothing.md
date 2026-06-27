# v353 Plan: Mobility Skill Smoothing

Date: 2026-06-26  
Spec: [`docs/specs/v353_spec-mobility-skill-smoothing.md`](../specs/v353_spec-mobility-skill-smoothing.md)

## Tasks

- [x] 1.1 Extend movement_presentation JSON + schema with mobility_smoothing
- [x] 1.2 Add MobilitySkillPresentation helper (leap/charge/teleport + travel reveal)
- [x] 1.3 Wire main.gd; suppress tick smoothing during active mobility
- [x] 1.4 Bot scenario 86 + wait/assert handlers + unit tests
- [x] 1.5 Docs: as-built, lifecycle, PROGRESS

## Verification

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_mobility_skill_presentation.gd
HEADLESS=1 make bot-client SCENARIO=86_mobility_skill_smoothing
```

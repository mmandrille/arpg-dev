# v353 Spec: Mobility Skill Smoothing

Status: Complete  
Date: 2026-06-26  
Codename: `mobility-skill-smoothing`  
Baseline: v352 `projectile-tick-smoothing`

## Purpose

Smooth client presentation for large mobility displacements: **leap**, **charge**, and **teleport**
(skill) for local and remote players, plus a brief teleport presentation when **teleporter floor
travel** reveals the player at a distant destination.

Extracts tween logic from `main.gd` into a focused presentation helper driven by shared tuning.

## Non-goals

- No server, protocol, shared golden, or sim changes.
- No earthbreaker/disengage mobility variants in v353.
- No projectile smoothing (v352).
- No dungeon torch lights (v354).

## Acceptance criteria

- `movement_presentation.v0.json` adds `mobility_smoothing` tuning (`enabled`,
  `teleport_duration_seconds`, `teleport_travel_duration_seconds`).
- `MobilitySkillPresentation` plays leap/charge/teleport for **local and remote** players on
  `skill_cast` events using authoritative landing positions from the same delta when available.
- Position tick smoothing is suppressed for entities with active mobility presentation.
- Teleporter travel plays a brief destination reveal when the post-travel position jump exceeds a
  rule-derived threshold.
- Bot debug exposes `mobility_skill_smoothing`; extended client bot scenario proves active-then-settled
  during a leap cast.

## Scope and likely files

| Area | Files |
|------|-------|
| Shared tuning | `movement_presentation.v0.json`, schema |
| Client helper | `mobility_skill_presentation.gd`, `movement_presentation_loader.gd` |
| Integration | `main.gd` |
| Bot | `86_mobility_skill_smoothing.json`, wait/assert handlers |
| Tests | `test_mobility_skill_presentation.gd` |

## Test and bot proof

```bash
make validate-shared
make client-unit
HEADLESS=1 make bot-visual scenario=86_mobility_skill_smoothing
```

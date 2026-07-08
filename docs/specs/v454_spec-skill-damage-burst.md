# v454 Spec — Skill Damage Burst Events

Status: Approved  
Date: 2026-07-08  
Codename: `skill-damage-burst`

## Purpose

Aggregate multi-hit skill damage into `skill_damage_burst` events; pilot hybrid instant-resolve for Magic Bolt while preserving client projectile flight from `skill_cast`.

## Acceptance

- Volley emits O(1) burst per cast on wire; bot assertions still pass via ingest expansion.
- Magic Bolt resolves on cast tick without projectile entity; `skill_cast` flight fields unchanged.
- `docs/adr/0016-combat-processing-budget.md` documents policy.
- Replay outcomes unchanged for pilot fixtures.

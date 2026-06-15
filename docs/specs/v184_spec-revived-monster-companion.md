# v184 Spec: Revived Monster Companion

Status: Complete
Date: 2026-06-15

## Goal

Add a Sorcerer Revive skill that turns a dead non-boss monster into a server-owned companion using the v182 companion behavior.

## Requirements

- Sorcerer has a `revive` skill in shared skill data.
- Revive targets a dead monster entity and rejects living targets.
- Revive explicitly rejects boss entities.
- Revived monsters become `companion` entities owned by the caster.
- Revived monsters keep the original monster definition/visual identity for client rendering.
- Rank 1 revive power is 50%, with +10% per additional rank for this slice.
- Revived companions follow the owner and attack nearby enemies using existing companion AI.
- Recast defaults to one active revived monster for now; v185 owns data-driven quantity limits.
- Bot proof kills a monster, revives the dead entity, and proves the revived companion damages another enemy.

## Non-Goals

- No persistence across sessions.
- No UI panel, commands, equipment, or leveling for revived companions.
- No boss revive.
- No multi-revive limit scaling; v185 owns that rule.

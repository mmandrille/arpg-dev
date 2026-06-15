# v189 Spec: Skill Affix Rolls

Status: Complete - `make ci` green on 2026-06-15

Codename: skill-affix-rolls

## Summary

Add higher-rarity skill utility affixes to the item roll system:
`skill_cooldown_reduction_percent` and `skill_mana_cost_reduction`, while keeping inherited
`all_skills` and `skill_damage_percent` live.

## Goals

- Add schema, validation, pricing, labels, and shared item-template candidates for cooldown and
  mana-cost skill affixes.
- Apply equipped cooldown reduction as a percent reduction to authoritative skill cooldowns.
- Apply equipped mana-cost reduction as a flat integer reduction to authoritative skill casts.
- Keep cooldowns and costs clamped to valid gameplay ranges: cooldown at least 1 tick, mana cost
  at least 0.
- Prove the behavior through focused Go tests and one compact protocol bot scenario.

## Non-goals

- No per-skill affix targeting, affix names, crafting, or skill tree UI redesign.
- No protocol version bump; existing skill cast and cooldown events already carry mana/cooldown data.
- No change to skill base definitions beyond item affix effects.

## Acceptance Criteria

1. Shared item schemas and validation accept `skill_cooldown_reduction_percent` and
   `skill_mana_cost_reduction`.
2. Rare-or-higher item templates can roll the new skill affixes, inherited by unique/set pools.
3. Equipped mana-cost reduction lowers skill cast mana spend and emitted `skill_cast.mana`.
4. Equipped cooldown reduction lowers `skill_cooldown_started.total_ticks`.
5. Existing `all_skills` and `skill_damage_percent` behavior remains covered.
6. `make maintainability`, `make validate-shared`, focused Go tests, the new bot scenario, and
   `make ci` pass.

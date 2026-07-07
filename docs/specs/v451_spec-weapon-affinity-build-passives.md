# v451 Spec - weapon-affinity-build-passives

## Goal

Make equipped active class affinities matter in the skill tree by letting passive skills scale from the number of active affinity rolls on worn gear.

## Scope

- Extend passive stat skills with optional affinity-count scaling owned by shared skill rules.
- Count active equipped class affinity rolls authoritatively in the server.
- Pilot the feature on two exemplar classes: `barbarian` and `rogue`.
- Add supporting affinity gear so each pilot can stack two active affinities in bot and gameplay labs.
- Add focused Go and bot proof.

## Non-goals

- Full six-class rollout.
- New affinity families.
- Market or respec wiring changes.
- Rebalancing every passive tree.

## Data / content decisions

- `unstoppable_heart` gains additional `damage_percent` from up to two active affinities.
- `killer_instinct` gains additional `crit_chance` from up to two active affinities.
- Added two pilot affinity gear templates:
  - `affinity_barbarian_girdle`
  - `affinity_rogue_treads`

## Adopt / borrow / reject

- Adopt: existing `class_affinities` gear payloads, passive skill stat path, skill progression lab, bot progression seeding.
- Borrow: v387 affinity item model/presentation patterns.
- Reject: protocol bump, external assets/plugins, and a broader class rollout in this slice.

## Proof

- Focused Go coverage for active/inactive affinity counting and passive scaling.
- Bot scenarios:
  - `make bot scenario=affinity_passive_barbarian`
  - `make bot scenario=affinity_passive_rogue`

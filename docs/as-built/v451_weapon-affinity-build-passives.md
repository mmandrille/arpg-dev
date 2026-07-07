# v451 As-Built - weapon-affinity-build-passives

## Shipped

- Extended `passive_stat_bonus` skills with optional `affinity_scaling` owned by shared skill rules.
- Added authoritative active-affinity counting across equipped gear.
- Piloted affinity-scaling passives on:
  - `unstoppable_heart` (barbarian)
  - `killer_instinct` (rogue)
- Added two supporting affinity gear templates:
  - `affinity_barbarian_girdle`
  - `affinity_rogue_treads`
- Extended `skill_progression_lab` and added two bot scenarios for live passive-allocation proof.

## Verification

- `make validate-shared` -> only the pre-existing unrelated `dungeon_teleporters` golden drift fails
- `cd server && go test ./internal/game -run 'TestClassAffinity|TestAffinityPassive'`
- `make bot scenario=affinity_passive_barbarian`
- `make bot scenario=affinity_passive_rogue`

## Deferred

- Sorcerer, paladin, ranger, and druid affinity-scaling passive rollout.
- Broader per-class affinity build rebalance after more affinity gear exists.

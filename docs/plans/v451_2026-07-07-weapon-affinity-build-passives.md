# v451 Plan - weapon-affinity-build-passives

## Implementation steps

1. Extend passive-skill rules/schema for optional affinity-count scaling.
2. Add an authoritative helper for counting active equipped class affinities.
3. Feed the count into passive-stat evaluation.
4. Author barbarian and rogue pilot passives on existing skill nodes.
5. Add two affinity gear templates plus skill lab loot/content.
6. Add focused Go coverage and two extended bot scenarios.
7. Update docs/progress and commit.

## Verification

- `make validate-shared` (known unrelated `dungeon_teleporters` golden drift may remain)
- `cd server && go test ./internal/game -run 'TestClassAffinity|TestAffinityPassive'`
- `make bot scenario=affinity_passive_barbarian`
- `make bot scenario=affinity_passive_rogue`

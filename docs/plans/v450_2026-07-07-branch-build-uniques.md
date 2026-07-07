# v450 Plan - branch-build-uniques

## Implementation steps

1. Extend reward/content data for named unique acquisition.
2. Add branch-skill unique effect definitions in shared rules.
3. Implement a server hook for skill-hit status unique effects.
4. Add four pilot uniques and route them through steward/boss/debug content.
5. Add focused Go coverage for unique execution and drop authoring.
6. Add extended bot scenarios for each pilot unique.
7. Update as-built/progress docs and commit.

## Verification

- `make validate-shared` (known unrelated `dungeon_teleporters` golden drift may remain)
- `cd server && go test ./internal/game -run 'TestBranchUnique|TestQuestStewardRewardFamilyCanGrantNamedUnique|TestBossTreasureClassUsesRandomEquipmentPool|TestBossLootRollsRandomEquipmentPayloads|TestUniqueBurn|TestOffensiveUnique|TestNamedUniquePayload'`
- `make bot scenario=branch_unique_warbrand`
- `make bot scenario=branch_unique_nightbloom`
- `make bot scenario=branch_unique_rimecoil`
- `make bot scenario=branch_unique_sunwake`

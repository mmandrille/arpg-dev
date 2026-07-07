# v450 As-Built - branch-build-uniques

## Shipped

- Added four pilot branch uniques:
  - `Warbrand Cleaver` (`gore_strike` -> burn rider)
  - `Nightbloom Shiv` (`death_blossom` -> poison rider)
  - `Rimecoil Staff` (`glacial_lance` -> splash slow)
  - `Sunwake Guard` (`divine_hammer` -> burn rider)
- Added `on_skill_hit_status` unique effect support in the server.
- Extended quest steward reward families to grant named uniques directly.
- Added one boss-floor treasure-class pin for `sunwake_guard`.
- Added `branch_unique_lab` and four extended bot scenarios.
- Fixed named unique payload authoring so `minimum_level` remains the authored equip floor after item-level stat scaling.
- Extended the bot harness so `equip_inventory_item` can select unique items by `display_name` and reports equip rejections with item context.

## Verification

- `make validate-shared` -> only the pre-existing unrelated `dungeon_teleporters` golden drift fails
- `cd server && go test ./internal/game -run 'TestBranchUnique|TestQuestStewardRewardFamilyCanGrantNamedUnique|TestBossTreasureClassUsesRandomEquipmentPool|TestBossLootRollsRandomEquipmentPayloads|TestUniqueBurn|TestOffensiveUnique|TestNamedUniquePayload'`
- `make bot scenario=branch_unique_warbrand`
- `make bot scenario=branch_unique_nightbloom`
- `make bot scenario=branch_unique_rimecoil`
- `make bot scenario=branch_unique_sunwake`

## Deferred

- Ranger and druid branch unique exemplars.
- Larger branch-unique catalog and broader boss/steward distribution.

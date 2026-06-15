# v180 Spec: Upgrade Resource Drop

## Intent

Introduce the first player-visible upgrade resource as loot. Item upgrades currently spend gold only;
this slice adds a tangible resource drop that can enter normal pickup, inventory, presentation, and
bot verification flows without adding a new wallet schema yet.

## Player-visible behavior

- `upgrade_shard` exists as a non-equippable currency item named "Upgrade Shard".
- The ranged lab multi-drop source drops an upgrade shard alongside its existing deterministic loot.
- The shard uses item presentation metadata so it has a distinct ground/inventory visual.
- Players pick it up explicitly like other non-gold loot, and it occupies inventory.

## Requirements

- Add `upgrade_shard` to item rules.
- Add presentation metadata for the shard.
- Add the shard to the deterministic `ranged_multi_drop` loot table.
- Add protocol bot coverage that kills the ranged dummy, picks up the shard, and asserts it is in
  inventory.
- Preserve existing gold auto-pickup behavior and consumable hotbar behavior.

## Non-goals

- Spending shards on upgrades.
- Adding an account-level resource wallet.
- Adding blacksmith UI resource counters.
- Retuning dungeon treasure-class probabilities.

## Verification

- `make validate-shared`
- Focused Go loot/drop tests as needed.
- `make bot scenario=72_upgrade_resource_drop.json`
- `make ci` before finish commit.

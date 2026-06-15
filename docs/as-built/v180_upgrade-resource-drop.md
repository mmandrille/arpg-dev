# v180 As-built — Upgrade Resource Drop

Date: 2026-06-15
Status: Complete

## What Shipped

- Added `upgrade_shard` as a non-equippable currency item named "Upgrade Shard".
- Added an `upgrade_resource` presentation family using a distinct box-shaped purple shard
  treatment, and mapped `upgrade_shard` to it.
- Added `upgrade_shard` to the deterministic `ranged_multi_drop` source so the ranged lab dummy
  drops it alongside gold, quest loot, and potions.
- Added protocol bot scenario `72_upgrade_resource_drop.json`, which kills the ranged dummy, picks
  up the upgrade shard, and asserts the shard is in inventory.

## Proof

- `make validate-shared`
- `make bot scenario=72_upgrade_resource_drop.json`
- `make bot scenario=06_ranged_lab.json`
- `make maintainability`
- `make ci`

## Follow-up Notes

- Shards are tangible loot only in this slice. Spending them on blacksmith upgrades still needs a
  dedicated resource-wallet or item-consumption contract.

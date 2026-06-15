# v180 Plan: Upgrade Resource Drop

## Scope

Add a tangible upgrade resource item to a deterministic monster drop and prove it can be picked up.

## Tasks

1. Shared data
   - Add `upgrade_shard` to `shared/rules/items.v0.json` as a currency, non-equippable item.
   - Add an `upgrade_resource` presentation family and map `upgrade_shard` to it.
   - Add `upgrade_shard` to `ranged_multi_drop` guaranteed drops and entries.

2. Bot proof
   - Add a new protocol bot scenario that reuses the ranged lab setup.
   - Kill `training_dummy_ranged`.
   - Pick up `upgrade_shard`.
   - Assert `upgrade_shard` exists in inventory and existing gold/consumable behavior still works.

3. Documentation and finish
   - Add an as-built note.
   - Update `PROGRESS.md` to v180 complete and mark the next step as the engineering review gate.
   - Run targeted validation, `make maintainability`, and `make ci`.

## Risks

- Adding another guaranteed loot entity could affect the ranged lab if pickup order or inventory
  expectations are too strict. The new scenario should assert the resource explicitly while keeping
  existing behavior unchanged.
- Spending shards is intentionally deferred because the current persistent account store only has
  gold and item rows, not a generic resource wallet.

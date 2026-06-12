# v103 As-Built: Unique Effect Catalog Foundation

Date: 2026-06-12

## What Shipped

- Added `shared/rules/unique_effects.v0.json` and its schema as the forward model for unique
  item behavior.
- Seeded three enabled global unique effect concepts:
  - `everburning_wound`: all hero damage applies burn for 10 seconds, ticking once per second for
    10% of the original hit damage.
  - `echoing_finish`: future low-health hit echo behavior.
  - `last_stand_glimmer`: future near-death equip passive behavior.
- Added validator cross-checks for unique-effect ids, enabled/ready status, supported hooks,
  item-type compatibility, and the burn tuning contract.

## Key Decisions

- Unique effects are global behavior hooks that can be attached to normal rolled equipment later.
  The old `unique_items.v0.json` seed remains disabled concept data, not the live runtime model.
- Burn tuning is shared data now so v105 can implement the live behavior without hardcoding the
  core cadence or percent.

## Verification

- `make validate-shared`
- `make maintainability`
- `make ci`

## Deferred

- Unique rarity rolls and item-instance effect attachment.
- Combat execution for unique effects.
- Client presentation, including the burning visual cue.

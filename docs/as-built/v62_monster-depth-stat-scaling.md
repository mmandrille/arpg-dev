# v62 As-Built - Monster Depth Stat Scaling

Date: 2026-06-11
Branch: `main`
CI: `make ci` green

## What Shipped

- Generated dungeon monsters now compute effective combat stats from base monster definition,
  dungeon depth, and rarity.
- Depth scaling covers HP, damage, armor, hit chance, crit chance, block percent, and attack
  cooldown, with rule-data caps/floors for volatile stats.
- Rarity definitions now add stat identity beyond loot/XP: champion, rare, and unique monsters can
  modify armor, hit chance, crit chance, block, and attack cadence.
- Static lab monsters and boss-template monsters keep their previous bespoke behavior; the new
  scaling is applied only during generated dungeon monster spawn.
- Monster rarity goldens and shared validation now assert effective stats for common, champion,
  rare, unique, and generated pinned cases.

## Regression Fixes

- Hotbar assignment persistence now treats paired `inventory_remove` plus `hotbar_update` changes
  as a view move instead of deleting the durable item. This preserves assigned consumables across
  fresh sessions while keeping the item out of the bag view.
- Bot scenarios were retuned where the broader combat pressure made old timing or random-hit
  assumptions brittle.

## Verification

- `make validate-shared`
- `make test-go`
- `make bot`
- `make ci`

# v450 Spec - branch-build-uniques

## Goal

Ship a first set of build-defining uniques that modify v420 branch/capstone skills instead of only adding raw stats.

## Scope

- Add four pilot named uniques for `barbarian`, `rogue`, `sorcerer`, and `paladin`.
- Author new unique effects that trigger from branch skills and apply v449 status effects.
- Route deterministic acquisition through the quest steward and one boss-floor treasure class pin.
- Surface the authored effect behavior in the existing item inspection/inventory flow.
- Add extended bot proofs for each unique.

## Non-goals

- Full unique catalog expansion.
- Ranger and druid branch exemplars.
- Crafting / blacksmith integration.
- New affix grammar or production art.

## Data / content decisions

- Reuse in-repo assets only.
- `warbrand_cleaver` modifies `gore_strike` with a burn rider.
- `nightbloom_shiv` modifies `death_blossom` with poison.
- `rimecoil_staff` modifies `glacial_lance` with splash slow.
- `sunwake_guard` modifies `divine_hammer` with burn.

## Adopt / borrow / reject

- Adopt: existing named unique payload path, quest steward reward families, unique chest debug flow, bot harness.
- Borrow: v449 status effect IDs and server-side debuff helpers.
- Reject: external art/plugins and a new protocol version.

## Proof

- Focused Go coverage for unique effect execution, quest steward named unique rewards, and boss treasure class authoring.
- Bot scenarios:
  - `make bot scenario=branch_unique_warbrand`
  - `make bot scenario=branch_unique_nightbloom`
  - `make bot scenario=branch_unique_rimecoil`
  - `make bot scenario=branch_unique_sunwake`

# v104 As-Built: Unique Drop Roll Contract

Date: 2026-06-12

## What Shipped

- Added `unique` as a server-loaded item rarity with its own prefix and four stat rolls.
- Loaded `shared/rules/unique_effects.v0.json` into Go rules.
- Updated the item roller so unique rarity attaches exactly one compatible unique effect id through
  the existing durable `effect_ids` payload field.
- Added item-roll golden coverage for a unique cave blade that carries `everburning_wound`.
- Added a compatibility test proving shield rolls cannot select the weapon-only
  `echoing_finish` effect.
- Updated generated shop-offer goldens for deterministic rarity-table drift while preserving the
  existing shop cap at rare.

## Key Decisions

- No protocol schema bump was needed because `effect_ids` already exists on loot, inventory, stash,
  and market item views.
- Shops remain capped at rare. The `unique` rarity is now rollable by the item roller, but vendor
  generated stock still rejects rolls above its configured cap.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game/...`
- `make maintainability`
- `make ci`

## Deferred

- Unique effect combat execution.
- Burn DOT ticks.
- Client burning visual cue.
- Bot-visual proof for a live unique effect.

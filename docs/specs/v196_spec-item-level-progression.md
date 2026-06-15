# v196 Spec: Item Level Progression

## Goal

Give rolled equipment an authoritative `item_level` that follows the loot source depth, so later upgrade and progression slices can reason about item power without inferring it from template names or requirements.

## Player-visible behavior

- Rolled equipment generated from a template carries an `item_level` field in floor loot, inventory, stash, shop, and appraisal protocol views.
- Item level is at least 1.
- For generated/template drops, item level equals the source depth already passed into the item roller.
- Named unique and set packages carry an item level equal to their effective minimum level.

## Scope

- Extend the durable rolled-item payload with `item_level`.
- Surface `item_level` in current v8 snapshot and delta schemas wherever rolled item metadata is already exposed.
- Add bot assertion support for checking rolled item levels.
- Add focused Go and bot coverage.

## Out of Scope

- Rebalancing item stats, rarity weights, or treasure-class weights.
- Adding client tooltip copy for item level.
- Changing equipment requirement rules.

## Acceptance Criteria

- A direct server test proves source depth `7` rolls item level `7`, while depth `0` clamps to `1`.
- Existing rolled loot transfer coverage proves floor loot and picked-up inventory views include the item level.
- A bot scenario picks up a rolled item and asserts `item_level`.
- `make validate-shared`, focused server tests, the bot scenario, and final `make ci` pass.

## Godot Plugin Adoption

Rejected for this slice. The change is server/protocol metadata only and does not introduce client UI, presentation, or placeholder art.

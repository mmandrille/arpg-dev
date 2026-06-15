# v179 Plan: Mana Regeneration

## Scope

Emit and validate an explicit passive mana regeneration event, then prove it through a bot scenario after real skill mana spend.

## Tasks

1. Protocol contract
   - Add `player_mana_regenerated` event requirements to latest state delta and session snapshot schemas.
   - Require `entity_id` and `mana`.

2. Server behavior
   - Update `applyPlayerRegen` to append `player_mana_regenerated` when mana regeneration adds at least 1 point.
   - Keep the existing entity update unchanged.
   - Leave health regen and consumable mana restoration untouched.

3. Tests
   - Extend the regen stat test to assert the new event appears only for mana regen ticks.
   - Keep expectations derived from existing rates and existing test timing.

4. Bot proof
   - Add `tools/bot/scenarios/71_mana_regeneration.json`.
   - Use debug progression to seed a mana-costing skill and enough magic to regenerate quickly.
   - Cast the skill, wait for `player_mana_regenerated`, and assert player mana has recovered above the post-cast value.

5. Documentation and finish
   - Add an as-built note.
   - Update `PROGRESS.md` to v179 and next v180.
   - Run targeted verification, then `make ci`, then commit `feat: v179: mana regeneration`.

## Risks

- Passive regen is tick/carry based, so bot timing must allow enough wait time for the configured magic-derived rate.
- The event should not reuse `player_mana_restored`, because potion restoration requires `item_instance_id` and has different semantics.

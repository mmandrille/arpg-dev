# v179 Spec: Mana Regeneration

## Intent

Make passive mana regeneration an explicit, end-to-end gameplay signal after a player spends mana. Mana already refills through derived character stats and equipment rolls; this slice closes the visibility gap by emitting a protocol event for passive mana gains and proving the loop with a bot scenario that spends mana, waits, and observes the refill without using a potion.

## Player-visible behavior

- When passive regeneration restores at least 1 mana, the server emits `player_mana_regenerated`.
- The event includes the player `entity_id` and restored `mana` amount.
- The normal player entity update still carries the authoritative current mana value, so existing bars and state consumers continue to update.
- Consumable mana restoration remains `player_mana_restored`; passive regeneration uses the new event type so tools can distinguish source.

## Requirements

- Add the event to the latest protocol schemas.
- Emit the event only when passive mana regeneration actually increases current mana.
- Do not emit the event while mana is full, dead, or the regen rate cannot produce an integer gain.
- Preserve existing health regen behavior.
- Add focused Go coverage for the new passive mana event.
- Add a protocol bot scenario that casts a mana-costing skill, waits for passive mana regeneration, and asserts the regenerated event and increased player mana.

## Non-goals

- Retuning base mana regeneration rates.
- Adding a new client visual effect.
- Changing potion mana restoration semantics.

## Verification

- `make validate-shared`
- Focused Go regen test package/run.
- `make bot scenario=71_mana_regeneration.json`
- `make ci` before finish commit.

# v179 As-built — Mana Regeneration

Date: 2026-06-15
Status: Complete

## What Shipped

- Added a distinct `player_mana_regenerated` protocol event for passive mana gains.
- Required `entity_id` and `mana` for that event in the v8 state delta and session snapshot
  schemas.
- Updated passive player regeneration so mana regen still sends the authoritative player entity
  update and now also emits the regen event when at least 1 mana is restored.
- Preserved potion restoration as `player_mana_restored`, keeping item-driven restoration separate
  from passive stat-driven regeneration.
- Added protocol bot scenario `71_mana_regeneration.json`, which casts Rage, waits for passive
  regen, and asserts mana rises without potion use.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run TestHealthAndManaRegenUseStatsAndItemRolls -count=1`
- `make bot scenario=71_mana_regeneration.json`
- `make maintainability`
- `make ci`

## Follow-up Notes

- No client-specific floating text or pulse was added. Existing client state updates still move the
  mana bar; the new event gives future UI work a clean source-specific hook.

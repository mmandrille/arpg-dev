# v105 Plan: Unique Burn Effect Live

## Scope

Implement live server behavior and client presentation for `everburning_wound`.

## Client shortcut decision

- **Reject addon adoption:** the visual cue is a small status marker and tint layered onto existing
- **Borrow pattern:** reuse the existing in-repo status marker pattern from
  `player_status_effect_markers.gd` and `main.gd` poison/slow tint handling.

## Tasks

- [x] Extend damage type support with `fire` across schema, validation, and Go canonicalization.
- [x] Add deterministic unique-effect DOT state to the sim snapshot lifecycle.
- [x] Trigger `everburning_wound` from equipped items after successful hero damage events.
- [x] Emit authoritative start/tick/end events using effect id `everburning_wound` and damage type
  `fire`.
- [x] Add focused Go tests for burn timing, damage, and unique item equipment activation.
- [x] Add client burning cue helpers and client tests/debug state assertions.
- [x] Add a protocol/client scenario proof for the live burn path.
- [x] Update `PROGRESS.md` and write as-built notes.
- [x] Run focused verification, `make maintainability`, and final `make ci`.

## Maintainability note

- `client/scripts/main.gd` and `server/internal/game/sim.go` exceeded their grandfathered
  line-count allowance by small integration hooks. Splitting the central delta/status dispatcher
  or player snapshot state during this slice would add more risk than value, so the baseline is
  intentionally updated for v105. New burn behavior is otherwise isolated in
  `server/internal/game/unique_effects.go` and `player_status_effect_markers.gd`.

## Verification targets

- `cd server && go test ./internal/game/... -run 'TestUniqueBurn|TestDamageType'`
- `make validate-shared`
- `make client-unit`
- `ARPG_BOT_SCENARIO=unique_burn_effect_live VERBOSE=1 make bot`
- `HEADLESS=1 make bot-client scenario=33_unique_burn_effect_live.json`
- `make maintainability`
- `make ci`

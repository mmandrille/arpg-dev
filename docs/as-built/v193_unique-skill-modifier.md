# v193 As-built: Unique Skill Modifier

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Added `on_skill_damage_roll` as a ready unique-effect hook for narrow skill-specific damage
  modifiers.
- Added the `Arcane Conduit` unique effect, which grants Magic Bolt 50% increased damage while
  equipped.
- Added the named unique `Conduit Staff` to the unique item catalog and debug unique chest payload.
- Threaded skill identity into the server skill damage path so matching unique effects can adjust
  the damage range before hit resolution.
- Kept basic attacks and non-target skills on the baseline damage path.
- Made unique chest tests generic over named uniques so future named packages do not require
  brittle special cases.
- Added protocol bot scenario `82_unique_skill_modifier.json`, proving the new named unique can be
  taken from the unique chest.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'UniqueSkill|UniqueChest|UniqueItemValidation' -count=1`
- `make bot scenario=82_unique_skill_modifier.json`
- `make ci`

## Deferred

- Non-damage skill modifiers such as projectile count, cooldown shape, or status payload changes.
- Client-specific presentation beyond existing unique tooltip/chest rows.

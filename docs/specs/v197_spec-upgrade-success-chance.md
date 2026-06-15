# v197 Spec: Upgrade Success Chance

## Goal

Make blacksmith item upgrades use a data-driven success chance so upgrade attempts can fail while still spending the configured gold cost.

## Player-visible behavior

- The blacksmith panel shows the configured upgrade success chance.
- Successful attempts behave as before: spend gold, increase item level, and improve one stat.
- Failed attempts spend gold, keep the item unchanged, and show a failed-attempt status instead of a generic error.

## Scope

- Add `item_upgrade_success_chance_percent` to `shared/rules/main_config.v0.json`.
- Validate the chance as an integer from 0 to 100.
- Thread the chance through the HTTP upgrade route into the atomic store transaction.
- Return `success` in upgrade responses.
- Update client blacksmith status/debug state and bot matcher support.
- Add focused store coverage for a forced failed attempt.

## Out of Scope

- Upgrade-resource consumption.
- Per-rarity or per-item chance curves.
- Pity counters or chance display styling polish.

## Acceptance Criteria

- Existing success-path upgrade tests remain deterministic at the default 100% chance.
- A forced 0% store test spends gold and leaves rolled stats unchanged.
- The blacksmith UI scenario asserts the default success chance is visible in panel state.
- `make validate-shared`, focused store/http/client checks, and final `make ci` pass.

## Godot Plugin Adoption

Rejected for this slice. The client work is a small existing-panel text/debug-state update and does not need new UI plugins, demos, or assets.

# Plan: `ui-currency-and-mana-polish`

Spec: [`docs/specs/v39_spec-ui-currency-and-mana-polish.md`](../specs/v39_spec-ui-currency-and-mana-polish.md)

## Tasks

- [ ] Update shared rules, protocol schemas, and examples for gold, mana, blue potions, and DEX armor.
- [ ] Add server-owned rolled/scaled gold pickup, gold persistence, and mana potion behavior.
- [ ] Update bot/tests/goldens from `training_badge` inventory items to gold or neutral bag fillers.
- [ ] Update Godot UI panels, settings, mana bar, gold counter, character rename affordance, and weapon mount pose.
- [ ] Validate with focused tests and `make ci`.

## Verification

- `make validate-shared`
- `go test ./internal/game/...`
- `make client-unit`
- `make ci`

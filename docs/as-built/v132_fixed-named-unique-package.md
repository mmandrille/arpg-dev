# v132 As-Built — Fixed Named Unique Package

Date: 2026-06-13

## What shipped

- Turned `embercall_blade` into an enabled ready named unique package.
- Added fixed rolled stats and fixed effect ids to named unique rule data.
- Loaded named uniques through Go rules and built deterministic `ItemRollPayload` rows from the
  base template plus fixed package fields.
- Extended the purple town unique chest so it grants Embercall Blade in addition to the generated
  enabled-effect coverage rows.
- Extended bot assertions so `61_purple_town_unique_chest` proves both coverage and named unique
  presence.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestNamedUnique|TestUniqueTestChest|TestLoadRules'`
- `.venv/bin/python -m pytest tools/bot/test_protocol.py -q`
- `make bot scenario=purple_town_unique_chest`
- `make maintainability`
- `make ci`

Manual visual check:

```bash
make bot-visual scenario=purple_town_unique_chest
```

## Deferred

- Natural named unique drop odds stay unchanged.
- Unique-market restrictions, additional named unique catalog entries, and player-facing unique
  effect inspection remain future slices.

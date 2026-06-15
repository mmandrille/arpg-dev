# v197 Plan: Upgrade Success Chance

## Spec

`docs/specs/v197_spec-upgrade-success-chance.md`

## Tasks

1. Extend main gameplay config/schema with `item_upgrade_success_chance_percent`.
2. Load and validate the chance in server rules.
3. Add success chance and roll inputs to store upgrade methods.
4. Keep upgrade spending atomic, but skip stat mutation when the roll fails.
5. Add `success` to stash and inventory upgrade HTTP responses.
6. Generate a per-attempt 1-100 success roll in the HTTP route.
7. Update blacksmith panel status, preview, debug state, and client-bot matcher support.
8. Add focused forced-failure store coverage and update existing success-path assertions.
9. Update as-built docs and `PROGRESS.md`, then run verification and commit.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/store ./internal/http -run 'Upgrade' -count=1`
- `make client-unit`
- `SCENARIO=blacksmith_upgrade_ui HEADLESS=1 ./scripts/bot_client_local.sh`
- `make ci`

## Notes

- Default success chance is 100%, preserving current player flow unless rules change.
- The forced-failure proof is at the store boundary because that is where spend and item mutation must stay atomic.

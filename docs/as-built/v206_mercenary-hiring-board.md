# v206 As-Built: Mercenary Hiring Board

Date: 2026-06-15
Status: Complete

## What shipped

- Added configurable `mercenary_hire_cost_gold` shared data with schema and Go validation.
- Added `town_mercenary_board` as a ready `mercenary` service interactable in town, `vendor_lab`, and a focused `mercenary_hiring_lab`.
- Added server-owned `action_intent` hiring from the board: opens the board offer, rejects unaffordable hires, spends gold on success, spawns one owned `mercenary_guard`, and persists wallet state.
- Added hired-mercenary replacement by source so repeated board hires do not stack unlimited guards.
- Added additive protocol event requirements for `mercenary_board_opened` and `mercenary_hired`.
- Added focused Go coverage for config validation, successful hire, insufficient gold, and replacement.
- Added `tools/bot/scenarios/88_mercenary_hiring_board.json` proving the board hire, companion follow, and companion damage path without expanding the bot runner.

## Proof

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestMercenaryHiring|TestMercenaryHireCost|TestMercenaryFoundation|TestCompanion'`
- `make bot scenario=mercenary_hiring_board`
- `make bot scenario=mercenary_foundation`
- `make ci`

## Notes

- Client presentation remains primitive/no-panel by design. The selected follow-up slice owns the roster UI.
- This is a fixed authored guard hire, not player-character-derived mercenary listing or snapshot persistence.

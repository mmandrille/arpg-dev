# v207 As-Built: Mercenary Roster UI

Date: 2026-06-16
Status: Complete

## What shipped

- Added a compact draggable `Mercenaries` panel that opens from `mercenary_board_opened`.
- The panel displays the fixed `fixed:mercenary_guard` offer, service id, price, current gold, affordability, status, and a filtered hired mercenary roster.
- Wired `mercenary_hired` to update the panel with the hired entity and current gold while reusing the same owned companion state that feeds `CompanionBar`.
- Added focused client-bot debug progression seeding so live Godot client scenarios can fund deterministic service interactions.
- Added mercenary panel wait/assert bot steps, with companion-bar count/icon assertions handled by a focused assertion helper.
- Added `tools/bot/scenarios/client/47_mercenary_roster_ui.json` to click the board, hire the guard, and prove the roster panel plus mercenary HUD icon.
- Added `client/tests/test_mercenary_panel.gd` and included it in `make client-unit`.

## Proof

- `make client-unit`
- `make bot-client scenario=mercenary_roster_ui`
- `make bot scenario=mercenary_hiring_board`
- `make maintainability`
- `make ci`

## Notes

- Server/protocol behavior is unchanged; v207 reflects the v206 board/hire events.
- The panel is still a fixed-offer roster view. Player-character mercenary listings, durable roster records, death rules, and stance commands remain deferred.

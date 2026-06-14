# v164 As-built — Session browser filters

Date: 2026-06-14
Status: Complete

## What shipped

- Added display-only search and sort controls to the Godot Join Game panel.
- Kept `GET /v0/sessions/active` and session join authority unchanged; filtering and sorting derive
  from the client-side active-session rows already returned by the server.
- Search matches visible session identity fields, including session id, host display name, world id,
  and mode.
- Sort modes cover recent update first, host name, and player count.
- Bot/debug state now reports total rows, filtered rows, visible rows, search text, and sort mode.

## Proof

- `make maintainability`
- `make client-unit`
- `make bot-client scenario=21_join_game_listed_session.json`
- `make ci`

Visual verification command:

```bash
make bot-visual scenario=21_join_game_listed_session.json
```

# v164 Spec — Session browser filters

Date: 2026-06-14
Status: Complete
Codename: session-browser-filters

## Purpose

Add display-only search and sort controls to the Godot Join Game session browser so listed co-op
sessions can be scanned before selecting a row.

The server remains authoritative for listed-session discovery and joins. The client filters and
sorts only the already returned active-session rows.

## Non-goals

- No backend query parameters, pagination, matchmaking, lobby invites, capacity rules, or Steam
  integration.
- No changes to session join authority, membership, visibility, or active-session storage.
- No party chat, ready checks, role labels, or richer party UI.

## Acceptance criteria

- Join Game panel exposes search text and sort controls.
- Filtering matches visible row identity such as host name, world id, and session id.
- Sorting supports recent first, host name, and player count without changing server data.
- Bot/debug state reports search text, sort mode, total row count, filtered row count, and visible
  session rows.
- Existing listed-session join flow remains green.

## Test and bot proof

- `make client-unit`
- `make bot-client scenario=21_join_game_listed_session.json`
- `make maintainability`
- `make ci`

Visual verification command:

```bash
make bot-visual scenario=21_join_game_listed_session.json
```

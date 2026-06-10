# v46 — Real Godot Join Game co-op proof

**Proves:** The player-facing Godot `Join Game` flow can join a real active listed co-op backend
session as a distinct guest account while a protocol-level host remains connected.

- Client bot preflight now creates a unique host account, creates a listed co-op `dungeon_levels`
  session, opens the host WebSocket, sends `client_ready`, and exposes the prepared session id to
  the Godot guest run.
- `scripts/bot_client.sh` supports scenario-scoped preflight setup and cleanup without changing
  normal non-preflight client-bot scenarios.
- `MultiplayerSessionsPanel` can select a specific session id for deterministic bot targeting.
- Client bot assertions can match active-session rows by expected id, verify current session id,
  and wait for remote-player presence through structured debug state.
- Client bot scenario `21_join_game_listed_session.json` proves root Join Game, active-list row
  visibility, Back-to-root behavior, selected listed join, co-op WebSocket metadata, and remote host
  presence in the real Godot client path.
- Existing v45 menu scenarios and protocol scenario `27_session_browser_uncapped_coop.json` remain
  green.

**Explicit non-goals:** no gameplay/protocol changes, Steam lobby, matchmaking, chat, ready checks,
filters/search/sorting, two-window Godot choreography, production lobby art/audio, or server model
change.

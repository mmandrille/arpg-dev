# v24 — Main menu and character start

**Proves:** The Godot client can boot into a player-facing shell while the server remains
authoritative for accounts, characters, ownership checks, and fresh-session bootstrap.

- Authenticated HTTP APIs list and create account-scoped named characters; duplicate display names
  are allowed in v24 and names are trimmed/length-limited server-side.
- Fresh session creation accepts an optional selected `character_id`, rejects cross-account
  character use, and preserves the default-character path for bots, smoke, replay, and dev flows.
- Interactive Godot startup now opens a main menu with Continue, New Game, Settings, and Exit;
  Continue starts a fresh `dungeon_levels` session from selected character progression.
- New Game creates a named character and starts a fresh `dungeon_levels` session for that
  character; old-world/session resume remains dev/debug-only.
- Local settings persist a fixed window size (`1280x720`, `1600x900`, `1920x1080`) in
  `user://settings.json` and apply immediately.
- ESC opens a pause menu with Resume, Settings, Return to Main Menu, and Exit; overlay visibility
  blocks gameplay clicks, WASD, hotbar, inventory, camera zoom, and bot-dispatched gameplay intents.
- Return to Main Menu closes the WebSocket, marks the session ended through a small idempotent
  owner-only route, clears client gameplay state, and offers only fresh character starts.
- Client bot scenario `08_main_menu_flow.json` proves settings, named character creation, pause
  input lock, return to menu, Continue, and fresh new session id; scenarios `01`-`07` keep their
  explicit auto-start path.

**Explicit non-goals:** no character delete/rename/class/customization/portraits, production menu
art/audio, richer settings, old-session resume UI, character summaries, stash/vendors/quests, or
durable dungeon maps/monsters/floor drops/HP/current level.

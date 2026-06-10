# v45 — Menu create/join flow

**Proves:** The player-facing menu now starts from backend-backed Create Game or Join Game flows
without implying offline play or old-session resume.

- Root menu actions are `Create Game`, `Join Game`, `Settings`, and `Exit`; `Continue`, `New Game`,
  and root `Multiplayer` are compatibility aliases for bots only.
- Settings persists a local `Create Game Type` preference with `Co-op` and `Solo`; `Co-op` remains
  the default and creates a listed co-op backend session.
- Create Game lists or creates characters, then starts a fresh backend `dungeon_levels` session as
  either listed co-op or solo based on the setting.
- Join Game opens the active listed-session browser first; character selection only appears after a
  selected listed session id exists.
- Character selection now exposes explicit choose-or-create and forced-create modes, keeps dead
  rows disabled, and preserves rename/delete affordances without reintroducing old root copy.
- Client debug state and bot actions now expose root menu labels/actions, character panel mode,
  create-game session type, selected join session id, and current session mode/listed flags.
- Client bot scenarios `08_main_menu_flow.json` and `20_menu_create_join_flow.json` prove the root
  menu, settings persistence, listed co-op create, solo create, Return to Main Menu, existing
  character reuse, and Join Game empty-state behavior.
- Protocol bot scenario `27_session_browser_uncapped_coop.json` remains the real backend listed
  discovery/join proof.

**Explicit non-goals:** no offline/local-only gameplay, player-facing old-session resume, Steam
lobbies/invites, matchmaking, chat, ready checks, filters/search/sorting, production menu art/audio,
or real multi-account Godot Join Game bot proof.

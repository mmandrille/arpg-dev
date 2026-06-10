# v38 — Session browser and uncapped co-op menu

**Proves:** Co-op can be hosted and joined from the main menu through server-listed sessions while
the authoritative session, realtime, replay, and persistence model supports more than two members.

- Sessions now persist a `listed` flag; authenticated `GET /v0/sessions/active` returns active listed
  co-op summaries without exposing join codes or account ids.
- Listed co-op creation and listed join are available through HTTP, Python bot helpers, and Godot
  `NetClient`; private co-op still uses the join-code path.
- The previous two-member cap is removed. Store, HTTP, realtime WebSocket, sim, and replay tests prove
  third and fourth joins, three connected clients, actor-scoped movement, distinct local player ids,
  same-level visibility, disconnect survival, and deterministic replay reconstruction.
- Godot adds a Multiplayer main-menu path with Host Listed Session, Refresh Sessions, Join Selected,
  and Back. Host/join flows reuse the character picker and connect through the existing WebSocket.
- Local `make play 3` now launches independent menu clients with distinct accounts and no pre-created
  co-op session. `BASE_URL=<url> make play-remote 3` launches menu clients against a remote backend
  after probing `/readyz`, without starting local DB/server.
- Protocol bot scenario `27_session_browser_uncapped_coop.json` proves listed discovery, two listed
  joins, three-peer visibility/movement, disconnect, reconnect, `/state`, and replay verification.

**Explicit non-goals:** no filters/search/sorting controls, Steam lobby/invites/friend flows, ready
checks, chat/emotes, party panel polish, trade, XP sharing, party bonuses, proximity reward rules,
loot allocation, PvP/friendly fire, load-aware capacity limits, split deployables, or cross-process
session ownership.

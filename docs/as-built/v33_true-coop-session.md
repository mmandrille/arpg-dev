# v33 — True co-op session

**Proves:** Two authenticated clients can join one Go-authoritative session, each controlling a
distinct character/player entity while replay, persistence, reconnect, and solo compatibility stay
deterministic.

- Protocol v2 schemas add `local_player_id`, `party[]`, actor metadata on player/combat/reward
  events, and actor-free client intents; v1 schemas remain intact.
- Sessions now support `mode: "coop"`, hashed join codes, deterministic `session_members`,
  per-member start snapshots, and actor-tagged input rows.
- HTTP create/join lets a host create a co-op session and a guest join by session id + join code;
  non-members are denied WebSocket access and duplicate member sockets are rejected.
- The sim now owns multiple player states with independent levels, inventories, hotbars,
  waypoints, progression, reconnect-to-town behavior, and non-solid player/player collision.
- A shared realtime session loop runs one authoritative sim per active session, binds each socket
  to its server-derived actor, sends recipient-scoped snapshots, and fans out level-visible deltas.
- Disconnecting a co-op member removes only that player's entity from other same-level clients;
  solo disconnects continue to preserve same-session resume behavior.
- Replay reconstructs host and late-joined guest members with actor-tagged inputs, member start
  snapshots, disconnect/reconnect state, and per-tick event sequence ordering.
- Godot stores `local_player_id`, keeps the local `PlayerAnchor` for camera/prediction/input, and
  renders other visible players as remote entity nodes with authoritative movement only.
- Protocol bot scenario `23_true_coop_session.json` proves host create, guest join, distinct local
  player ids, party metadata, same-level visibility, independent movement, guest disconnect/removal,
  guest reconnect to town, and replay verification.

**Explicit non-goals:** matchmaking/lobby, public discovery, Steam lobby/invites, party panel
polish, chat/emotes/ready checks, trade, XP sharing, party bonuses, loot allocation rules,
friendly fire/PvP, production remote-player art, more than two players, and distributed session
ownership across server processes.

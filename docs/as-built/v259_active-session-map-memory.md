# v259 As-built - Active-Session Map Memory

Date: 2026-06-18

## What shipped

- Added an active session key to `DiscoveryMinimap`.
- `DiscoveryMinimap.sync_session(session_key)` resets explored cells only when a non-empty session
  key changes.
- `DiscoveryMinimap.reset_session()` clears the key and explored state during gameplay teardown.
- Snapshot application syncs the minimap session key from the existing snapshot/client session id.
- Explored cells persist while the same session remains active, including compact/full-screen map
  mode changes.
- Bot debug state exposes `session_key`, and `assert_discovery_minimap` can assert that the key is
  present without pinning the dynamic session id.
- Added `72_active_session_map_memory` as a focused client scenario.

## Proof

```bash
make client-unit
HEADLESS=1 make bot-visual scenario=72_active_session_map_memory
make maintainability
```

## Manual visual check

```bash
make bot-visual scenario=72_active_session_map_memory
```

## Scope limits

- No server, shared protocol, replay, persistence, database, reconnect/resume, or cross-session
  character map memory shipped.
- No durable fog snapshots, server-owned explored-cell model, marker changes, biome changes, LOS
  changes, or dungeon-generation changes shipped.

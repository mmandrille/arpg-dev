# v259 Plan - Active-Session Map Memory

Status: Complete
Goal: Scope discovery-map explored memory to the active client play session.
Architecture: Keep memory client-presentational. Add a session key to `DiscoveryMinimap`, reset the
state tracker on key changes and gameplay teardown, and leave reconnect/resume/cross-session memory
out of scope. No protocol or server work.
Tech stack: Godot GDScript client, Godot client bot scenario, docs.

## Baseline and Shortcut Decision

Builds on v256-v258 discovery minimap state, modes, and POI markers. `session_snapshot` already
contains `session_id`, and `NetClient` tracks the active session locally.

Asset/plugin decision: reject external assets/plugins. This is lifecycle/state plumbing in the
existing minimap scripts only.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/discovery_minimap.gd` | Add active session key tracking, reset, and debug field |
| Modify | `client/scripts/main.gd` | Sync minimap session key from snapshots and reset on teardown |
| Modify | `client/tests/test_discovery_minimap.gd` | Prove reset on session change and retention within a session |
| Add | `tools/bot/scenarios/client/72_active_session_map_memory.json` | Client proof for session key and mode-change retention |
| Modify | `docs/specs/v259_spec-active-session-map-memory.md` | Mark complete during close-out |
| Modify | `docs/progress/slice-lifecycle.md` | Add v259 lifecycle row |
| Add | `docs/as-built/v259_active-session-map-memory.md` | Record shipped behavior and proof |
| Modify | `PROGRESS.md` | Update current status and next selected autoloop item |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none planned
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Keep state logic in `DiscoveryMinimap`; touch `main.gd` only at snapshot/teardown integration.

Verification:
```bash
make maintainability
```

## Task 1 - Session-Scoped Minimap State

Files:
- Modify: `client/scripts/discovery_minimap.gd`
- Modify: `client/tests/test_discovery_minimap.gd`

- [x] Step 1.1: Add `sync_session(session_key)` and `reset_session()` APIs.
- [x] Step 1.2: Reset explored state when a non-empty session key changes.
- [x] Step 1.3: Preserve explored state when syncing the same session repeatedly.
- [x] Step 1.4: Expose `session_key` in debug state.
- [x] Step 1.5: Add unit tests for same-session retention and session-change reset.

```bash
make client-unit
```

## Task 2 - Main Integration and Bot Proof

Files:
- Modify: `client/scripts/main.gd`
- Add: `tools/bot/scenarios/client/72_active_session_map_memory.json`

- [x] Step 2.1: Sync the minimap session key when applying snapshots.
- [x] Step 2.2: Reset minimap session memory when gameplay state tears down.
- [x] Step 2.3: Add a client scenario that asserts a non-empty session key and explored count
  survives compact/full-screen mode changes in the same session.

```bash
HEADLESS=1 make bot-visual scenario=72_active_session_map_memory
```

## Task 3 - Lifecycle Docs

Files:
- Modify: `docs/specs/v259_spec-active-session-map-memory.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v259_active-session-map-memory.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Mark the v259 spec complete.
- [x] Step 3.2: Add v259 lifecycle and as-built notes.
- [x] Step 3.3: Update `PROGRESS.md` current status and leave biome, door, LOS, and quest marker
  work as remaining selected autoloop scope.

```bash
make maintainability
```

## Final Verification

- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-visual scenario=72_active_session_map_memory`
- [x] `make maintainability`
- [ ] Autoloop final batch gate: `make ci`

Manual visual proof, if desired:

```bash
make bot-visual scenario=72_active_session_map_memory
```

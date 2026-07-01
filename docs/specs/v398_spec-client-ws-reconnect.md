# v398 Spec — Client WebSocket Reconnect

Status: Complete
Date: 2026-07-01
Codename: client-ws-reconnect
Baseline: v397 `item-archetype-library` complete

## Purpose

When the gameplay WebSocket drops while an interactive session is active and the server session
still exists, the Godot client enters a **Reconnecting…** state, retries with exponential backoff,
reattaches to the same `session_id`, sends `client_ready` with `last_seen_tick`, applies the
authoritative `session_snapshot`, and resumes play without returning to the main menu.

## Decisions (defaults applied)

| # | Decision |
|---|----------|
| Q-1 | WS redial to same `session_id` first; HTTP `create_session(resume_session_id)` fallback after repeated WS failures |
| Q-2 | Client-only slice; accept solo in-memory loop rebuild on reconnect |
| Q-3 | Co-op members may respawn in town per existing server `session_loop.attach` behavior |
| Q-4 | **Cancel → main menu** available on overlay after the first failed attempt |
| Q-5 | Pause menu blocked while reconnect overlay is active |
| Q-6 | Backoff tuning in `shared/rules/main_config.v0.json` → `client_reconnect` block |

## Non-goals

- Server solo session-loop grace period after last socket detach
- Menu-driven “resume old session” browser (ADR-0008 deferred)
- Token refresh / production auth recovery
- New protocol envelope types or schema version bumps
- Merge-gate bot scenario simulating real network partitions

## Acceptance criteria

- [ ] After the first `session_snapshot`, an unintended WS `CLOSING`/`CLOSED` during `gameplay_active` shows a blocking **Reconnecting…** overlay
- [ ] Gameplay intents are blocked while reconnecting; pause menu cannot open
- [ ] Retry uses data-driven exponential backoff from `client_reconnect` in `main_config.v0.json`
- [ ] Success path: WS redial → `client_ready` → `session_snapshot` resync → overlay hides and input works
- [ ] After repeated WS failures, HTTP `create_session(resume_session_id)` fallback then WS connect
- [ ] Give-up shows **Connection lost** with **Return to menu** (calls existing teardown)
- [ ] Cancel appears after first failed attempt and returns to main menu
- [ ] Intentional disconnect (`_return_to_main_menu`, `_exit_game`, new session) does not trigger auto-reconnect
- [ ] Recovery logic extracted from `main.gd` (`connection_recovery.gd`); ratchet respected
- [ ] Godot unit tests for backoff/state machine; `make client-unit` green; `smoke.gd` resume path still green

## Scope and files

| Area | Files |
|------|-------|
| Shared | `shared/rules/main_config.v0.json`, `main_config.v0.schema.json` |
| Client | `connection_recovery.gd`, `connection_overlay.gd`, `main_config_loader.gd`, `net_client.gd`, `main.gd` |
| Tests | `client/tests/test_connection_recovery.gd`, `scripts/client_smoke.sh` registration |
| Docs | plan, as-built, lifecycle |

## Test and bot proof

```bash
make validate-shared
make client-unit
make client-smoke
```

Server reconnect remains covered by existing `ws_test.go` and bot `check_persistence`.

## Asset decision

Reject external UI assets; reuse `level_loading_overlay.gd` layout patterns for code-native overlay.

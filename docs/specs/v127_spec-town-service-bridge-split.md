# v127 Spec - Town Service Bridge Split

Status: Complete
Date: 2026-06-13
Codename: `town-service-bridge-split`

## Purpose

Reduce `client/scripts/main.gd` ownership by moving town-service inventory-context wiring into a
focused bridge helper. Market and blacksmith panels should continue to open with the inventory panel
in the right mode, and double-click inventory staging should still route to the active service.

## Non-goals

- No server/API changes.
- No market, blacksmith, stash, or inventory behavior changes.
- No new UI design.
- No broad `main.gd` refactor beyond town-service bridge wiring.

## Acceptance Criteria

1. A dedicated Godot helper owns market/blacksmith inventory-context bridge operations.
2. `main.gd` delegates market context toggles, blacksmith context toggles, and inventory staging
   intents to the helper.
3. Existing market/blacksmith inventory staging behavior remains covered by tests.
4. A focused headless GDScript unit test covers the bridge helper.
5. Client smoke and CI pass.

## Likely Files

- `client/scripts/town_service_bridge.gd`
- `client/scripts/main.gd`
- `client/tests/test_town_service_bridge.gd`
- `scripts/client_smoke.sh`
- `docs/as-built/v127_town-service-bridge-split.md`
- `PROGRESS.md`

## Test Proof

```bash
godot --headless --path client --script res://tests/test_town_service_bridge.gd
make client-smoke
make ci
```

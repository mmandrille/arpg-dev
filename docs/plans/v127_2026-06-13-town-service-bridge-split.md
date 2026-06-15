# v127 Plan - Town Service Bridge Split

Status: Complete
Goal: Extract market/blacksmith inventory-context routing from `main.gd` into a focused helper.
Architecture: `main.gd` remains the orchestrator for server actions and state mutation. A stateless
`TownServiceBridge` helper owns panel context toggles and inventory staging intent routing.
Tech stack: Godot GDScript, headless client tests, SDD docs.

## Baseline And Shortcut Decision

Baseline is v126 `skill-validation-split` on `main`, committed as `77073168`.
Note: the worktree contains unrelated user changes around client settings, market refresh, and
server test cleanup. This slice preserves those changes and stages only v127-owned files.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/town_service_bridge.gd` | Market/blacksmith inventory context and staging routing. |
| Modify | `client/scripts/main.gd` | Delegate bridge operations to the helper. |
| Add | `client/tests/test_town_service_bridge.gd` | Focused helper proof. |
| Modify | `scripts/client_smoke.sh` | Run the focused bridge test in smoke. |
| Add | `docs/as-built/v127_town-service-bridge-split.md` | Record implementation and proof. |
| Modify | `PROGRESS.md` | Lifecycle closeout. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files likely touched:
- [x] `client/scripts/main.gd`
- [x] `scripts/client_smoke.sh`

Decision:
- [x] Extract only bridge glue; leave market/blacksmith action flows in `main.gd` for a later
  service-action split.

## Task 1 - Bridge Helper

- [x] Step 1.1: Add helper functions for opening/closing market inventory context.
- [x] Step 1.2: Add helper functions for opening/closing blacksmith inventory context.
- [x] Step 1.3: Add helper function for routing inventory staging intents to market/blacksmith
  panels.

## Task 2 - Main Wiring And Tests

- [x] Step 2.1: Preload the helper in `main.gd`.
- [x] Step 2.2: Replace inline context and staging logic with helper calls.
- [x] Step 2.3: Add headless helper tests and wire them into `client_smoke.sh`.

## Task 3 - Lifecycle

- [x] Step 3.1: Run focused test and client smoke through the new bridge gate.
- [x] Step 3.2: Run isolated `make ci` for v127-owned changes.
- [x] Step 3.3: Mark docs complete, update `PROGRESS.md`, and commit.

## Final Verification

- [x] `godot --headless --path client --script res://tests/test_town_service_bridge.gd`
- [x] `make client-smoke` reached and passed `GDScript town service bridge test`; final live smoke
  failed because no server was running on `:18081`.
- [x] `make ci`

# v127 As-Built - Town Service Bridge Split

Date: 2026-06-13
Spec: [`docs/specs/v127_spec-town-service-bridge-split.md`](../specs/v127_spec-town-service-bridge-split.md)
Plan: [`docs/plans/v127_2026-06-13-town-service-bridge-split.md`](../plans/v127_2026-06-13-town-service-bridge-split.md)

## What Shipped

- Added `client/scripts/town_service_bridge.gd` for market/blacksmith inventory context and staging
  intent routing.
- `main.gd` delegates town-service bridge operations while keeping server action handling in place.
- Added `client/tests/test_town_service_bridge.gd` and wired it into `scripts/client_smoke.sh`.

## Proof

- `godot --headless --path client --script res://tests/test_town_service_bridge.gd`
- `make ci`

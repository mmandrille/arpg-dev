# v67 Plan: Boss Kill Reward Polish

## Spec

[`docs/specs/v67_spec-boss-kill-reward-polish.md`](../specs/v67_spec-boss-kill-reward-polish.md)

## File Map

- `server/internal/game/types.go` — add optional `boss_template_id` to authoritative events.
- `server/internal/game/sim.go` — emit `boss_killed` for boss deaths without changing generic
  kill, loot, XP, or unlock order.
- `server/internal/game/game_test.go` — assert boss kill events include the new boss event.
- `shared/protocol/session_snapshot.v*.schema.json` and `state_delta.v*.schema.json` — allow and
  require `boss_template_id` on `boss_killed`.
- `shared/protocol/examples/state_delta.json` — document the event shape.
- `client/scripts/main.gd` — display boss defeat status from `boss_killed`.
- `tools/bot/scenarios/24_boss_floor_gate.json` — assert the new protocol event.
- `tools/bot/scenarios/client/28_boss_phase_readability.json` — assert the status message after
  the existing boss phase readability proof.
- `PROGRESS.md`, `docs/as-built/v67_boss-kill-reward-polish.md` — close-out documentation.

## Tasks

- [x] Add and test the server `boss_killed` event.
- [x] Update shared protocol schemas/examples for the event payload.
- [x] Add Godot status feedback and client-bot assertion.
- [x] Update protocol bot boss-floor proof.
- [x] Update docs and run focused verification plus `make ci`.

## Verification

```bash
make validate-shared
(cd server && go test ./internal/game)
make bot scenario=24_boss_floor_gate.json
HEADLESS=1 make bot-client scenario=28_boss_phase_readability.json
make ci
```

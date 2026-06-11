# v67 As-built: Boss Kill Reward Polish

## What shipped

- Boss deaths now emit a dedicated `boss_killed` event alongside the existing generic
  `monster_killed` event.
- `boss_killed` carries the boss entity id, source/target ids, and `boss_template_id`, giving
  future reward hooks and clients a stable boss-specific signal without changing loot, XP, combat,
  or exit-unlock behavior.
- The Godot client stores and exposes a boss reward status such as `Cave Warden defeated`; the
  visible status text includes it when status text is enabled.
- Protocol and Godot client bot coverage now prove the Cave Warden boss kill event and client
  reward status.

## Proof

- `make validate-shared`
- `(cd server && go test ./internal/game)`
- `make bot scenario=24_boss_floor_gate.json`
- `HEADLESS=1 make bot-client scenario=28_boss_phase_readability.json`
- `make ci`

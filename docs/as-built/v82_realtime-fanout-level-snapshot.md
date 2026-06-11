# v82 As-built - Realtime Fanout Level Snapshot

## What shipped

The realtime session loop now snapshots each connected client's current level while holding
`sessionLoop.mu` during `doTick`. `fanoutResult` consumes that captured `playerID -> level` map
instead of querying `Sim.PlayerCurrentLevel` after the mutex is released.

## What it proves

- Tick persistence and realtime fanout now share the same tick-time client-level view.
- Existing same-level and cross-level fanout behavior remains intact.
- Direct fanout tests pass explicit level snapshots, making the concurrency boundary visible in
  focused coverage.
- The Magic Bolt, Rage, Heal, Holy Shield, and matching client bot scenarios no longer pin
  attack-interval-derived exact values that are already owned by shared golden tests.
- The model-reaction client scenario now uses a safe low-HP dummy in `combat_stat_lab` for the
  terminal monster reaction proof and waits for any authoritative player damage for the local hit
  reaction.

## Verification

```bash
cd server && go test ./internal/realtime/...
VERBOSE=1 make bot scenario=32_skill_points_and_magic_bolt
VERBOSE=1 make bot scenario=39_rage_and_heal_skills,40_paladin_heal_skill
VERBOSE=1 make bot scenario=43_paladin_holy_shield
make validate-shared
VERBOSE=1 HEADLESS=1 make bot-client scenario=12_model_reaction_polish
make maintainability
make ci
```

## Deferred

- Broader realtime loop extraction.
- Protocol or gameplay behavior changes.
- New player-facing bot scenarios; none were needed for this internal fanout boundary.

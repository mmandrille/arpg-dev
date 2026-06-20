# v298 Spec - Barbarian Leap Obstacle Crossing

Status: Complete
Date: 2026-06-19
Codename: barbarian-leap-obstacle-crossing

## Purpose

Let the Barbarian `leap` mobility skill jump across marked floor obstacles, starting with water and
holes, while ordinary walking and non-Leap mobility still treat those tiles as movement blockers.
The behavior should be server-authoritative, rules-owned, and visible through the existing Leap arc
presentation.

This builds on v295 water, v296 holes, and v297 obstacle-kind plumbing. Leap becomes the first player
mobility exception without introducing general player swimming, flying, falling, or bridge systems.

## Non-goals

- No normal walking, click-to-move, player auto-navigation, Dash, Charge, Teleport, projectile,
  monster, companion, loot placement, fog/LOS, or dungeon generation behavior changes.
- No falling damage, rescue, landing recovery, bridge placement, terrain cost, swim state, or
  persistent traversal abilities.
- No protocol schema change; existing `skill_cast` and entity update output remain sufficient.
- No new imported assets, shaders, Godot plugins, or Leap visual timing changes.

## Acceptance Criteria

- `shared/rules/skills.v0.json` marks Barbarian `leap` with schema-backed obstacle kinds it may
  cross: `water` and `hole`.
- Server validation rejects unsupported mobility obstacle exceptions, especially `wall`.
- Leap endpoint resolution ignores water/hole while sweeping its jump path but still requires a
  valid floor landing; if the range ends inside water/hole, Leap lands on the nearest valid floor
  before that obstacle instead of standing in it.
- Leap still stops before normal walls and closed interactable barriers.
- Dash/Charge and ordinary player walking remain blocked by water and holes.
- A compact preset lab proves water/hole obstacles are present and gives Leap enough distance to
  cross them without touching production dungeon generation.
- Focused Go tests prove Leap crosses water/hole, refuses to land inside them, still stops at walls,
  and Dash/ordinary movement do not inherit the exception.
- A protocol bot scenario casts Leap across the lab and asserts the player lands beyond the
  water/hole strip.
- A headless visual bot replay for the same scenario shows the existing Leap arc over the rendered
  water/hole strip.

## Scope and Likely Files

- Shared skill/world rules:
  - `shared/rules/skills.v0.json`
  - `shared/rules/skills.v0.schema.json`
  - `shared/rules/worlds.v0.json`
- Server rules/mobility:
  - `server/internal/game/rogue_rules.go`
  - `server/internal/game/mobility_skills.go`
  - `server/internal/game/rogue_skills.go`
  - `server/internal/game/mobility_obstacle_crossing_test.go`
- Bot proof:
  - `tools/bot/scenarios/100_barbarian_leap_obstacle_crossing.json`
- Docs:
  - `docs/plans/v298_2026-06-19-barbarian-leap-obstacle-crossing.md`
  - `docs/as-built/v298_barbarian-leap-obstacle-crossing.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets, shaders, and plugins. Borrow the existing Leap visual
event path and the existing water/hole rendering introduced by v295-v297.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'LeapObstacle|MobilityObstacle|RogueDash|GeneratedObstacleCollisionPaths'`
- `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=barbarian_leap_obstacle_crossing ./scripts/bot_local.sh`
- `make client-unit`
- `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=barbarian_leap_obstacle_crossing ./scripts/bot_visual.sh`
- `make maintainability`

Manual visual proof, if desired:

```bash
make bot-visual scenario=barbarian_leap_obstacle_crossing
```

## Open Questions and Risks

- No required questions for this run. Defaults: only `leap` receives the exception, the exception is
  data-owned on the skill, and landing inside water/hole remains disallowed.
- Risk: current Leap endpoint code is shared with Dash/Charge helpers. Keep the implementation
  explicit so the exception cannot leak into Rogue Dash or Barbarian Charge.
- Risk: exact Leap landing depends on step resolution. Tests should assert semantic positions
  beyond/before obstacle bands rather than pinning brittle floating point values.

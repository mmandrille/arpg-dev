# ADR-0008: World Structure & Dungeon Progression

- **Status:** Accepted
- **Date:** 2026-06-06
- **Deciders:** Project owner (PM / tech lead)
- **Context tags:** action-RPG, dungeon, progression, co-op, procedural-generation, character-persistence

---

## Context

The project has a working authoritative server + thin Godot client foundation (slices v0–v17).
All gameplay so far occurs in a single, pre-defined world layout per session. Inventory and
progression are session-scoped — they do not persist when the session ends.

This ADR establishes the **game world structure and progression loop**: the town hub, the
infinite inverted-tower dungeon, procedural level generation, waypoints, the co-op model, and
the shift to character-scoped persistence. These decisions will shape many future slices and must
be agreed upon before any content or progression work begins.

### Narrative premise

A traveller arrives at a town whose central church has been consumed by a rift. A cave has opened
in the ruins; monsters are pouring out and attacking the town. The town is the safe hub — vendors,
quest-givers, and a blacksmith operate there as NPC entities. The dungeon is an **infinite inverted
tower** that descends beneath the ruined church: every level down is harder, procedurally generated,
and populated with stronger monsters and better loot.

---

## Decisions

### D1 — Character-scoped persistence replaces session-scoped

**Decision:** A `characters` Postgres table holds all cross-session character state: inventory,
equipped items, gold, stats, skills, waypoints unlocked, and deepest level reached. Sessions
reference a character; session-scoped inventory is removed.

**Why now:** The dungeon is infinite and meaningful progression requires state that survives
session boundaries. Waypoints (see D4) only make sense if the character remembers which levels
it has reached.

**Rejected:** keeping session-scoped inventory and bolting on a separate character table later —
the refactor cost grows with every slice that assumes the current model.

### D2 — The `Sim` holds a `levels map[int]*LevelState`

**Decision:** The Go `Sim` struct is extended to hold a map of concurrent level states keyed by
level number. Level 0 is town. Levels -1, -2, … are dungeon floors. All levels visited during a
session remain alive for its full lifetime — monsters, loot on the floor, dead bodies, and open
doors are preserved exactly as the player left them.

The deterministic 20 Hz tick processes all active `LevelState` instances each tick. Entity IDs
remain globally unique across levels within a session (monotonic counter, same policy as ADR-0001
D8). The existing single-world Sim is the direct predecessor; `LevelState` encapsulates what
`Sim` currently holds for one world.

**Why:** Fits directly onto the existing architecture. Replay remains deterministic — the input
stream records which level each intent targets; reconstruction replays all level states in the
same order. Co-op is natural: two players are entities in whichever `LevelState` they currently
occupy within the same Sim.

**Memory note:** all visited levels stay loaded for the session lifetime. For a session that
descends to level 50, all 50 `LevelState` instances live in memory. This is acceptable for the
current scale. If sessions regularly exceed ~50 active levels, a level snapshot/eviction upgrade
(serialize inactive levels to Postgres, deserialize on return) is the natural follow-on — it does
not require changing the `LevelState` interface.

**Rejected:**
- *Per-level sub-sessions* — breaks deterministic replay; cross-level co-op coordination becomes
  a separate coordination problem.
- *Eager eviction now* — premature; adds serialization complexity before the problem exists.

### D3 — Levels are generated on-demand, seeded

**Decision:** When a player first descends to level N in a session, the server generates it using
`session_seed + abs(N)` as the PCG seed. Generation parameters (room count, corridor style,
monster density curve, loot density curve) are declared in `shared/rules/` as data, not code —
both Go and any future tooling consume the same generation catalog.

A level's PCG output is fully deterministic from its seed: same seed always produces the same
map layout. This means session replay reconstructs levels correctly without storing map data.

**Between sessions:** character's `waypoints_unlocked` persists (see D4), but map data does not.
A new session has a fresh `session_seed`, so returning to level -5 via waypoint generates a new
map for that level — the character arrives at a fresh dungeon floor.

### D4 — Waypoints persist on the character; maps do not

**Decision:** Each dungeon level contains exactly one waypoint interactable. Activating it (via
the existing `action_intent` flow) writes the level number to `character.waypoints_unlocked` in
Postgres. Town always has an active waypoint — it is the default fast-travel hub.

Using a waypoint opens a menu of unlocked destinations. Travel is instant: the player entity
moves to the target `LevelState`, which is generated on-demand if not yet visited this session.

**What persists cross-session:** waypoint access (which level numbers the character can jump to).
**What does not persist cross-session:** map layout, monster positions, loot on the floor — all
regenerate from the new session seed.

### D5 — Town is level 0 — a static pre-defined world

**Decision:** Town is `level 0` in the `levels` map. It is a static world layout defined in
`worlds.v0.json` (the existing mechanism), not procedurally generated. Town is a safe zone: combat
intents are rejected at level 0. NPC entities (vendors, quest-givers, blacksmith) are entity types
that live permanently in the town `LevelState` for the session's lifetime.

The rift/cave entrance at the ruined church is an interactable that triggers descent to level -1.
Dungeon stairs (down/up) are interactables handled by `descend_intent` / `ascend_intent`.

### D6 — Co-op: multiple players share the same Sim and see each other per level

**Decision:** Multiple player entities coexist in the same `Sim` across its `LevelState` map.
Players on the same level share the same `LevelState` — they see each other, fight the same
monsters, and can interact (trade, assist). The server sends `state_delta` envelopes scoped to
each player's current level; a player on level -3 receives deltas for level -3 only, including
all other players present there.

Transitioning between levels (descend, ascend, waypoint) moves the player entity from one
`LevelState` to another within the same Sim. Players on different levels are invisible to each
other, which is the correct behavior — they are physically elsewhere in the dungeon.

**Deferred:** multiplayer session creation, matchmaking, and trade protocol design are deferred
to future ADRs. This ADR establishes only that the architecture must support co-op and that the
multi-level Sim model accommodates it without structural changes.

### D7 — Quest structure (framework only)

**Decision:** Every 3 dungeon levels, a town NPC offers a quest tied to that depth band (e.g.
"bring me X from level -3 through -5"). Quest progress and completion state are character-scoped
and persist across sessions. Completing a quest rewards a notable item delivered to character
inventory. Quest NPCs are entity types in the town `LevelState`.

**Deferred to a future ADR:** quest dialog protocol, objective type catalog (kill N, collect X,
reach level Y), reward delivery mechanics, and multi-quest concurrency.

---

## Protocol changes required

| Change | Reason |
|--------|---------|
| `state_delta` envelope gains a `level` field | Client must know which level a delta belongs to |
| New intents: `descend_intent`, `ascend_intent`, `use_waypoint_intent` | Level transitions and fast travel |
| Session snapshot includes current level and active level states | Resume and reconnect correctness |

These are additive protocol changes and require a schema version bump in `shared/protocol/`.

---

## Consequences

### Immediate (future slices must implement)

- `characters` table: new Postgres schema; sessions reference a character
- `Sim` refactor: `world *World` → `levels map[int]*LevelState`; `LevelState` encapsulates current per-world state
- PCG foundation: `shared/rules/dungeon_generation.v0.json` with generation parameters; Go evaluator seeded per `session_seed + abs(level)`
- Level transition intents and protocol delta scoping
- Town as static `level 0` with safe-zone combat rejection
- Waypoint interactable type and `character.waypoints_unlocked` persistence

### Deferred

- Quest system design (dialog, objective types, reward delivery) — future ADR
- NPC entity interaction protocol — future ADR/spec
- Player trade — future ADR
- Character progression formulas (stats, skills, leveling curves) — future ADR/spec
- Level snapshot/eviction for very deep sessions — only if memory becomes a real problem
- Multiplayer session creation and matchmaking — future ADR

---

## Relationship to existing ADRs

| ADR | Relationship |
|-----|-------------|
| [0001](0001-technology-stack.md) | D2 (authoritative server), D6 (shared rules), D8 (determinism) — all preserved; this ADR extends the Sim model without violating them |
| [0006](0006-asset-pipeline.md) | Unaffected; asset pipeline applies equally to dungeon and town assets |
| [0007](0007-animation-state-model.md) | Unaffected; animation remains client-only and event-driven regardless of level count |

---

## As-built addendum — World Detail & Navigation (v302–v308)

The v302–v308 batch extended the generated world model under this ADR (D3 PCG, server
authority). Recorded here because these are world-structure decisions ADR-0008 owns; they fit the
existing model rather than replacing it. All tuning is data in `shared/rules/`; the server owns every
outcome and the client renders only.

- **Obstacle `kind` taxonomy.** Generated walls now carry a `kind`: `wall`, `water`, `hole`, `rock`,
  `column`, `rubble`. Water and holes are hard-blocking floor hazards; `rock`/`column`/`rubble` are
  solid variety obstacles. Weights live in `dungeon_generation.obstacle_generation.solid_kind_weights`;
  water/hole sizing and counts in `obstacle_generation.water` / `.holes`. Placement is reachability-
  validated (no unreachable down-stair). Server: `dungeon_floor_features.go`,
  `dungeon_obstacle_variety.go`, `obstacle_blocking.go`. (v302, v303, v306)
- **Per-monster navigation trait.** `monsters.v0.json` gains `navigation_trait` (`grounded` | `flying`);
  flying monsters ignore water/hole blocking in both pathfinding and live movement via one shared
  predicate (`monster_navigation_traits.go`, `sim.go`). (v304)
- **Skill obstacle-crossing.** Mobility skills may declare `ignore_obstacle_kinds` (e.g. barbarian leap
  ignores `water`/`hole`); the leap sweep stops at hard walls and rejects landing inside an ignored
  obstacle (`mobility_skills.go`, `skills.v0.json`). (v305)
- **Line-of-sight-gated fog.** Walls may set `blocks_line_of_sight`; the server computes monster
  visibility through fog and only reveals what the player can see (`fog_of_war.go`). Render metadata
  (`kind`, `blocks_line_of_sight`) is additive/optional on the existing `wall` protocol def — no version
  bump, backward-compatible. (v300/v307)

Deferred (still future ADR/slice work): non-rectangular/polygon LoS occlusion, destructible/secret
obstacles, boss-floor obstacle generation, true flying gameplay/pathing beyond ground-ignore, and final
biome/difficulty balance — see PROGRESS.md "Dungeon generation" open gaps.

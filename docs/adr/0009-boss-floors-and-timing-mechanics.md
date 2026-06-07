# ADR-0009: Boss Floors & Timing Mechanics

- **Status:** Proposed
- **Date:** 2026-06-07
- **Deciders:** Project owner (PM / tech lead)
- **Context tags:** action-RPG, dungeon, boss, procedural-generation, combat, telegraph, progression-gate

---

## Context

[ADR-0008](0008-world-structure-and-dungeon-progression.md) establishes the infinite inverted-tower dungeon:
levels are procedurally generated on-demand from a seeded PCG pipeline, with up/down stairs,
waypoints, and monster placement driven by `shared/rules/dungeon_generation.v0.json`.

Current dungeon floors (v18–v21) are flat arenas with random monster spawns and a training-badge
loot drop on non-entry floors. There is no milestone pacing, no treasure chest interactable, and
no combat pattern beyond chase-and-melee (`dungeon_mob` with `attack_cooldown_ticks`).

The dungeon needs **rhythm**: periodic set-piece encounters that reward observation and timing
rather than pure gear checks. Boss floors provide that rhythm and gate descent until the player
proves they can read telegraphed attacks.

This ADR defines **when** boss floors occur, **how** they are laid out in the PCG pipeline, **how**
procedurally assembled bosses fight using authoritative timing mechanics, and **how** progression
through the floor is gated.

---

## Decisions

### D1 — Every 5 dungeon levels is a boss floor

**Decision:** A level is a **boss floor** when `levelNum < 0` and `abs(levelNum) % 5 == 0`.
Boss floors are: **-5, -10, -15, -20, …** Level -1 (entry hall) is never a boss floor.

**Why:** A fixed cadence gives players predictable milestone pacing without requiring hand-authored
content per depth. Five levels aligns with ADR-0008 D7's quest depth bands (every 3 levels) while
keeping boss spacing wide enough that normal floors still matter.

**Rejected:**
- *Boss on every level ending in 5 only at shallow depths* — adds special cases; the modulo rule
  scales cleanly to infinite depth.
- *Random boss chance* — breaks learnability and deterministic replay; agents cannot assert a
  fixed layout from seed alone.

### D2 — Boss-floor layout order: chest → boss → locked down stairs

**Decision:** Boss-floor PCG produces a **fixed narrative sequence** on the critical path from
up-stair entry to depth progression:

```text
[ up stairs ] → … optional trash mobs … → [ treasure chest ] → [ boss arena ] → [ down stairs (locked) ]
```

| Element | Rule |
|---------|------|
| **Treasure chest** | Exactly one `treasure_chest` interactable, always placed on the path between the up-stair approach zone and the boss spawn. Opened via existing `action_intent`; rolls loot from a depth-scaled boss-tier loot table in `shared/rules/`. |
| **Boss** | Exactly one boss entity in a dedicated arena cell cluster near the down-stair side of the floor. Trash mob count on boss floors is reduced (data-driven in `dungeon_generation.v0.json`). |
| **Down stairs** | Spawn at the normal PCG down-stair position but with `initial_state: "locked"`. Unlock when the boss entity is killed (see D6). Up stairs and waypoint placement follow the standard non-boss algorithm. |

Placement uses the same seeded RNG stream as normal floors (`session_seed + "|" + abs(N)` per
ADR-0008 D3). The chest, boss, and down stairs are placed in that order with minimum-separation
constraints so the player always encounters them in sequence.

**Why:** The user-facing promise is "reward, then gate, then depth." Chest-before-boss gives a
tangible payoff before the fight; locked down stairs make the boss the explicit progression
checkpoint.

**Rejected:**
- *Chest after boss* — removes pre-fight reward tension and encourages skip-running to stairs.
- *Optional chest off the critical path* — violates the always-after-chest guarantee.

### D3 — Boss identity is procedurally assembled from a data catalog

**Decision:** Bosses are not hand-placed per level. Each boss floor selects a **boss template**
from `shared/rules/boss_templates.v0.json` using the level's seeded RNG:

```json
{
  "template_id": "cave_warden",
  "base_monster_def_id": "dungeon_mob",
  "pattern_deck": ["ground_slam", "line_sweep", "summon_adds"],
  "hp_multiplier": 8,
  "loot_table": "boss_drop_tier_1",
  "asset_id": "boss_cave_warden"
}
```

Selection inputs: `abs(levelNum)` depth band (which template pool is eligible), level seed (which
template within the pool), and optional cosmetic variant index. The assembled boss is a
`monsterEntity` with `is_boss: true`, elevated `max_hp`, a reference to its active pattern deck,
and standard loot-on-death behavior.

Stat scaling (HP multiplier, damage modifier) is declared in rules data keyed by depth band so
both Go and future tooling share one catalog. Same seed always produces the same boss on the same
level — replay reconstructs the fight without storing boss assembly separately.

**Why:** Procedural assembly keeps infinite depth viable for AI-built content while preserving
determinism. Templates reuse the existing monster entity pipeline instead of a parallel boss type.

**Rejected:**
- *Fully random stat rolls per spawn* — harder to balance and to golden-test; template + multiplier
  is sufficient variance.
- *Separate `bossEntity` kind* — unnecessary; boss behavior is monster AI + pattern state machine.

### D4 — Timing mechanics are authoritative, tick-phased, and telegraph-first

**Decision:** Boss attacks follow **attack patterns** declared in
`shared/rules/boss_patterns.v0.json`. Each pattern is a ordered list of **phases** with durations
in sim ticks (20 Hz, same clock as everything else):

| Phase kind | Authoritative behavior | Player-facing purpose |
|------------|----------------------|------------------------|
| `telegraph` | Boss enters windup; hitboxes inactive; aim direction / zone geometry fixed for this phase | Client shows telegraph VFX (floor decal, windup anim) — player learns where/when to move |
| `active` | Damage hitboxes or projectiles apply; positions resolved server-side | Player must be out of zone (or blocking, when added) or take damage |
| `recovery` | Boss vulnerable; no pattern advance attacks; optional reduced move speed | Reward window for player attacks |

Example pattern (`ground_slam`):

```json
{
  "pattern_id": "ground_slam",
  "phases": [
    { "kind": "telegraph", "duration_ticks": 30, "telegraph_shape": "circle", "radius": 3.0 },
    { "kind": "active",   "duration_ticks": 4,  "damage": { "min": 4, "max": 6 }, "shape": "circle", "radius": 3.0 },
    { "kind": "recovery", "duration_ticks": 20 }
  ],
  "cooldown_ticks": 40
}
```

The boss AI state machine (server-only Go) cycles: **idle/agro → pick pattern from deck → execute
phases tick-by-tick → cooldown → repeat**. Pattern selection uses the level-local combat RNG stream
(labeled sub-sequence; must not consume unrelated `Sim.rng` rolls — same policy as dungeon PCG).

**Telegraph guarantee:** Every `active` phase that deals damage must be preceded by a `telegraph`
phase in the same pattern. No damage without windup. Minimum telegraph duration floor
(e.g. 20 ticks = 1 s) is enforced at rules validation time.

**Why:** Timing combat is the core design goal. Authoritative phases preserve ADR-0001 D2 (server
owns outcomes), ADR-0001 D8 (determinism), and give agents replayable, assertable fight timelines.
Data-driven patterns let AI agents author new bosses without Go changes.

**Rejected:**
- *Client-side hit detection for boss zones* — violates authoritative boundary.
- *Wall-clock telegraphs* — breaks replay; ticks only.
- *Instant undodgeable attacks mixed into patterns* — violates telegraph-first guarantee.

### D5 — Boss phase state crosses the wire as events, not animation

**Decision:** Following [ADR-0007](0007-animation-state-model.md), boss windups and strikes are
**not** animation states on the wire. The server emits gameplay events in `state_delta.events`:

| Event | Payload (conceptual) | Client use |
|-------|---------------------|------------|
| `boss_phase_started` | `entity_id`, `pattern_id`, `phase_index`, `phase_kind`, `aim`, `zone` | Start telegraph VFX / windup clip |
| `boss_phase_ended` | `entity_id`, `pattern_id`, `phase_index` | End telegraph, transition anim |
| `player_damaged` / existing | (unchanged) | Damage feedback when active phase hits |

Optional: include a compact `boss_pattern` snapshot on the boss entity in level-scoped deltas
(current pattern, phase index, ticks remaining) so reconnect/resume is correct without replaying
events. Animation mapping stays client-only in `main.gd`.

**Why:** Same proven split as monster hit/death — server emits facts, client renders them.

### D6 — Down stairs unlock on boss death

**Decision:** On boss floors, `stairs_down` starts in a new interactable state `locked`.
`descend_intent` targeting locked stairs is rejected with a `descend_blocked` event (reason:
`boss_alive`).

When the boss entity transitions to dead (`monster_killed` or a dedicated `boss_killed` event if
we need to distinguish trash from boss), the server sets the down-stair interactable state to
`ready`. All players on that level can then descend.

Co-op: one shared boss kill unlocks stairs for everyone on the level — same policy as shared
monster state in ADR-0008 D6.

**Why:** Makes the boss an explicit progression gate without a separate quest flag.

**Rejected:**
- *Key item from chest required to unlock stairs* — extra inventory friction; chest loot is reward,
  not key.
- *Per-player unlock* — inconsistent with shared `LevelState`.

### D7 — Golden fixtures and bot scenarios for at least one boss pattern

**Decision:** The first implementation slice must include:

- A golden fixture (`shared/golden/boss_floor_-5.json`) asserting PCG layout: chest position,
  boss template ID, locked stairs, unlock after kill.
- A golden fixture for one full pattern timeline (phase tick boundaries + damage ticks) cross-checked
  in Go and Godot headless tests.
- A Python bot scenario that opens the chest, survives one telegraphed pattern via timed move, kills
  the boss, and descends — proving agent-playability.

**Why:** Boss timing is exactly the class of bug that replay and bots catch; deferring tests would
violate project invariants.

---

## Protocol changes required

| Change | Reason |
|--------|--------|
| New events: `boss_phase_started`, `boss_phase_ended` | Telegraph/active phase facts for client VFX |
| Optional: `boss_killed`, `descend_blocked` | Clear progression and rejection feedback |
| Interactable state `locked` for `stairs_down` | Boss-floor gate |
| New interactable def: `treasure_chest` | Boss-floor reward container |
| Entity snapshot fields for boss pattern progress (optional) | Reconnect/resume correctness |
| Schema version bump in `shared/protocol/` | Additive changes |

---

## Shared rules additions

| File | Purpose |
|------|---------|
| `shared/rules/boss_templates.v0.json` | Template catalog: pattern decks, scaling, loot, assets |
| `shared/rules/boss_patterns.v0.json` | Phase sequences, telegraph shapes, damage zones |
| `shared/rules/dungeon_generation.v0.json` | `boss_floor` block: trash density, placement constraints, depth bands |
| `shared/rules/interactables.v0.json` | `treasure_chest`, `stairs_down.locked` state |
| `shared/rules/loot_tables.v0.json` | Boss-tier chest and kill drops per depth band |

All pattern and template logic is **data, not code** (ADR-0001 D6).

---

## Consequences

### Immediate (future slices must implement)

- Boss-floor branch in `GenerateDungeonLevel` (detect `abs(N) % 5 == 0`)
- `treasure_chest` interactable + open/action flow + loot roll
- Boss template selection + scaled monster spawn
- Boss pattern state machine in Go sim (tick-phased, deterministic)
- Locked down stairs + unlock on boss death
- Protocol events + client telegraph presentation (decals/windup anims)
- Golden fixtures + bot scenario

### Deferred

- Multi-phase boss enrage below HP threshold — future ADR/spec
- Co-op boss scaling (HP/damage per player count) — future ADR
- Block/parry interaction with active phases — future combat ADR
- Distinct boss arena geometry (walls/partitions) vs open floor — start with open floor + zone decals
- Boss health bar UI — client polish slice
- Procedural boss **appearance** beyond template `asset_id` swap — art pipeline slice

---

## Relationship to existing ADRs

| ADR | Relationship |
|-----|--------------|
| [0001](0001-technology-stack.md) | D4 preserves authoritative server + determinism; patterns are shared rules data |
| [0007](0007-animation-state-model.md) | D5 — telegraph presentation is client-only; server emits phase events |
| [0008](0008-world-structure-and-dungeon-progression.md) | Extends D3 PCG and level interactables; boss floors are a generation mode, not a new level type |
| [0006](0006-asset-pipeline.md) | Boss `asset_id` entries follow existing manifest + mount conventions |

---

## Open questions (for review before Accepted)

1. **Depth band loot tiers:** Should chest loot and boss kill loot share one table or separate
   (`boss_chest_drop` vs `boss_kill_drop`)?
2. **Trash mobs during boss fight:** Should remaining trash mobs pause aggro while boss is active,
   or stay active for added chaos?
3. **Minimum telegraph duration:** Is 1 s (20 ticks) the right floor for all patterns, or should
   shallow bosses (-5) telegraph longer than deep bosses (-50)?
4. **Waypoint on boss floors:** Standard waypoint placement remains, or waypoint moved to
   post-boss side only?
